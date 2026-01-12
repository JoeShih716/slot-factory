package wallet

import (
	"log/slog"
	"time"

	"github.com/shopspring/decimal"
)

type PaymentError struct {
	Code    int
	Message string
}

// TransactionRecord 代表一筆錢包交易紀錄。
type TransactionRecord struct {
	ID              int64           `json:"id"`
	PlayerID        string          `json:"playerID"`
	Amount          decimal.Decimal `json:"amount"`
	TransactionType string          `json:"transactionType"`
	BalanceAfter    decimal.Decimal `json:"balanceAfter"`
	CreatedAt       time.Time       `json:"createdAt"`
}

// HistoryProvider 定義了讀取交易歷史的介面，用於 API 服務層。
type HistoryProvider interface {
	GetHistory(playerID string, limit int) ([]TransactionRecord, *PaymentError)
}

// Payment 定義了完整錢包含業務邏輯介面（包含寫入）。
type Payment interface {
	HistoryProvider
	// GetBalance 取得玩家餘額。
	GetBalance(playerID string) (balance decimal.Decimal, err *PaymentError)

	// Debit 扣款。
	// 如果餘額不足，返回錯誤。
	Debit(playerID string, amount decimal.Decimal) (newBalance decimal.Decimal, err *PaymentError)

	// Credit 加款。
	Credit(playerID string, amount decimal.Decimal) (newBalance decimal.Decimal, err *PaymentError)

	// DebitAndCredit 扣款和加款。
	DebitAndCredit(playerID string, debitAmount decimal.Decimal, creditAmount decimal.Decimal) (newBalance decimal.Decimal, err *PaymentError)
}

type Service struct {
	payment Payment
	logger  *slog.Logger
}

// NewService 建立一個新的錢包服務實例。
func NewService(logger *slog.Logger, payment Payment) *Service {
	return &Service{
		payment: payment,
		logger:  logger.With("component", "wallet_service"),
	}
}

func (s *Service) GetBalance(playerID string) (decimal.Decimal, *PaymentError) {
	balance, err := s.payment.GetBalance(playerID)
	if err != nil {
		s.logger.Error("get balance failed", "playerID", playerID, "error", err)
		return decimal.Zero, err
	}
	return balance, nil
}

func (s *Service) Debit(playerID string, amount decimal.Decimal) (decimal.Decimal, *PaymentError) {
	newBalance, err := s.payment.Debit(playerID, amount)
	if err != nil {
		s.logger.Error("debit failed", "playerID", playerID, "amount", amount, "error", err)
		return newBalance, err
	}
	return newBalance, nil
}

func (s *Service) Credit(playerID string, amount decimal.Decimal) (decimal.Decimal, *PaymentError) {
	newBalance, err := s.payment.Credit(playerID, amount)
	if err != nil {
		s.logger.Error("credit failed", "playerID", playerID, "amount", amount, "error", err)
		return decimal.Zero, err
	}
	return newBalance, nil
}

func (s *Service) DebitAndCredit(playerID string, debitAmount decimal.Decimal, creditAmount decimal.Decimal) (decimal.Decimal, *PaymentError) {
	newBalance, err := s.payment.DebitAndCredit(playerID, debitAmount, creditAmount)
	if err != nil {
		s.logger.Error("debit and credit failed", "playerID", playerID, "debitAmount", debitAmount, "creditAmount", creditAmount, "error", err)
		return newBalance, err
	}
	return newBalance, nil
}

func (s *Service) GetHistory(playerID string, limit int) ([]TransactionRecord, *PaymentError) {
	records, err := s.payment.GetHistory(playerID, limit)
	if err != nil {
		s.logger.Error("get history failed", "playerID", playerID, "error", err)
		return nil, err
	}
	return records, nil
}
