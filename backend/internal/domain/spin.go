package domain

import "github.com/shopspring/decimal"

// Status 定義了 spin 的不同狀態
type Status int

const (
	Normal Status = iota
	HitFreeGame
	FreeGame
	FreeGameHitFreeGame
)

// GameResult 定義了特定遊戲結果所需實現的介面
type GameResult interface {
	MarshalJSON() ([]byte, error)
	UnmarshalJSON([]byte) error
}

// SpinResult 代表一次 spin 的結果。
type SpinResult struct {
	// ID 表示該次 spin 的唯一識別 ID。
	ID uint64 `json:"id"`
	// GameID 代表這次 spin 所屬的遊戲 ID。
	GameID int `json:"gameID"`
	// Result 代表這次 spin 的結果，可以是任何類型，視遊戲規則而定。
	Result GameResult `json:"result"`
	// BetAmount 代表這次 spin 下注的總金額。
	BetAmount decimal.Decimal `json:"betAmount"`
	// WinAmount 代表這次 spin 贏得的總金額。
	WinAmount decimal.Decimal `json:"winAmount"`
	// Status 代表這次 spin 的狀態，例如是否進入免費遊戲等。
	Status Status `json:"status"`
}
