package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/joe_shih/slot-factory/internal/application/wallet"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// TransactionModel 對應資料庫的 wallet_transactions 表，用於紀錄流水。
type TransactionModel struct {
	ID              int64           `gorm:"primaryKey;autoIncrement"`
	PlayerID        string          `gorm:"column:player_id"`
	Amount          decimal.Decimal `gorm:"column:amount;type:decimal(18,4)"`
	TransactionType string          `gorm:"column:transaction_type"`
	BalanceAfter    decimal.Decimal `gorm:"column:balance_after;type:decimal(18,4)"`
	CreatedAt       time.Time       `gorm:"column:created_at"`
}

func (TransactionModel) TableName() string {
	return "wallet_transactions"
}

// ProxyPayment 實現了 wallet.Payment 介面，
// 它會呼叫外部 API 並將成功的異動紀錄寫入本地 DB (Audit Log)。
type ProxyPayment struct {
	db      *gorm.DB // 用於紀錄流水，如果 db 為 nil 則跳過紀錄
	baseURL string
	apiKey  string
	client  *http.Client
}

var _ wallet.Payment = (*ProxyPayment)(nil)

// NewPayment 建立一個新的 Proxy 錢包實作。
func NewPayment(db *gorm.DB, baseURL, apiKey string) *ProxyPayment {
	return &ProxyPayment{
		db:      db,
		baseURL: baseURL,
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

// --- 介面實作 ---

func (p *ProxyPayment) GetBalance(playerID string) (decimal.Decimal, *wallet.PaymentError) {
	// 這裡示範呼叫外部 API
	resp, err := p.callAPI("GET", fmt.Sprintf("/balance/%s", playerID), nil)
	if err != nil {
		return decimal.Zero, &wallet.PaymentError{Code: 502, Message: "External API error"}
	}
	return resp.Balance, nil
}

func (p *ProxyPayment) Debit(playerID string, amount decimal.Decimal) (decimal.Decimal, *wallet.PaymentError) {
	reqBody := map[string]interface{}{
		"playerID": playerID,
		"amount":   amount,
	}
	resp, err := p.callAPI("POST", "/debit", reqBody)
	if err != nil {
		return decimal.Zero, &wallet.PaymentError{Code: 502, Message: "External API error"}
	}

	// 寫入本地流水紀錄
	p.logTransaction(playerID, amount.Neg(), "BET", resp.Balance)

	return resp.Balance, nil
}

func (p *ProxyPayment) Credit(playerID string, amount decimal.Decimal) (decimal.Decimal, *wallet.PaymentError) {
	reqBody := map[string]interface{}{
		"playerID": playerID,
		"amount":   amount,
	}
	resp, err := p.callAPI("POST", "/credit", reqBody)
	if err != nil {
		return decimal.Zero, &wallet.PaymentError{Code: 502, Message: "External API error"}
	}

	// 寫入本地流水紀錄
	p.logTransaction(playerID, amount, "PAY", resp.Balance)

	return resp.Balance, nil
}

func (p *ProxyPayment) DebitAndCredit(playerID string, debitAmount, creditAmount decimal.Decimal) (decimal.Decimal, *wallet.PaymentError) {
	reqBody := map[string]interface{}{
		"playerID":     playerID,
		"debitAmount":  debitAmount,
		"creditAmount": creditAmount,
	}
	resp, err := p.callAPI("POST", "/spin", reqBody)
	if err != nil {
		return decimal.Zero, &wallet.PaymentError{Code: 502, Message: "External API error"}
	}

	// 寫入本地流水紀錄 (紀錄淨額)
	p.logTransaction(playerID, creditAmount.Sub(debitAmount), "BETANDPAY", resp.Balance)

	return resp.Balance, nil
}

func (p *ProxyPayment) GetHistory(playerID string, limit int) ([]wallet.TransactionRecord, *wallet.PaymentError) {
	if p.db == nil {
		return nil, &wallet.PaymentError{Code: 500, Message: "Local database not enabled"}
	}

	var models []TransactionModel
	err := p.db.Where("player_id = ?", playerID).Order("created_at DESC").Limit(limit).Find(&models).Error
	if err != nil {
		return nil, &wallet.PaymentError{Code: 500, Message: "Database error"}
	}

	records := make([]wallet.TransactionRecord, len(models))
	for i, m := range models {
		records[i] = wallet.TransactionRecord{
			ID:              m.ID,
			PlayerID:        m.PlayerID,
			Amount:          m.Amount,
			TransactionType: m.TransactionType,
			BalanceAfter:    m.BalanceAfter,
			CreatedAt:       m.CreatedAt,
		}
	}
	return records, nil
}

// --- 輔助方法 ---

type apiResponse struct {
	Balance decimal.Decimal `json:"balance"`
	Status  string          `json:"status"`
}

func (p *ProxyPayment) callAPI(method, path string, body interface{}) (*apiResponse, error) {
	var bodyReader *bytes.Buffer
	if body != nil {
		jsonData, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(jsonData)
	}

	url := p.baseURL + path

	var finalReader *bytes.Buffer
	if body != nil {
		finalReader = bodyReader
	}

	req, _ := http.NewRequest(method, url, nil)
	if finalReader != nil {
		req, _ = http.NewRequest(method, url, finalReader)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	// 這裡目前只是個示意，因為還沒有真的外部 API
	// 您之後可以根據對手的 API 文檔來精修此處格式
	_ = req // 展示用

	// Mock 回傳 (之後請換成真實 http 呼叫)
	return &apiResponse{
		Balance: decimal.NewFromInt(100000),
		Status:  "ok",
	}, nil
}

func (p *ProxyPayment) logTransaction(playerID string, amount decimal.Decimal, txType string, balanceAfter decimal.Decimal) {
	if p.db == nil {
		return
	}
	// 非同步寫入流水，不要擋住遊戲主邏輯
	go func() {
		err := p.db.Create(&TransactionModel{
			PlayerID:        playerID,
			Amount:          amount,
			TransactionType: txType,
			BalanceAfter:    balanceAfter,
		}).Error
		if err != nil {
			// 如果流水沒記成，建議至少要噴個 Error Log 便於對帳
			fmt.Printf("[ALARM] Local transaction log failed: %v\n", err)
		}
	}()
}
