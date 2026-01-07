package game1001

import "github.com/shopspring/decimal"

// ActionType 定義了此遊戲內部使用的訊息動作類型。
type ActionType string

const (
	ActionPlayerJoined ActionType = "player_joined"
	ActionPlayerLeft   ActionType = "player_left"
	ActionPlayerList   ActionType = "player_list"
	ActionStateUpdate  ActionType = "state_update"
	ActionPlayerBet    ActionType = "player_bet"
	ActionOpening      ActionType = "opening"
	ActionWinResult    ActionType = "win_result"
	ActionBetResult    ActionType = "bet_result"
)

// --- Payloads ---

// PlayerInfo 定義了廣播給前端的玩家資訊。
type PlayerInfo struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	BetAmount decimal.Decimal `json:"betAmount"`
}

// PayloadPlayerList 是發送給新玩家的當前玩家列表。
type PayloadPlayerList struct {
	Players []PlayerInfo `json:"players"`
}

// PayloadPlayerJoined 是廣播給房間內所有人的新玩家資訊。
type PayloadPlayerJoined struct {
	Player PlayerInfo `json:"player"`
}

// PayloadPlayerLeft 是廣播給房間內所有人的離開玩家資訊。
type PayloadPlayerLeft struct {
	PlayerID string `json:"playerId"`
}

// PayloadStateUpdate 廣播遊戲狀態變更。
type PayloadStateUpdate struct {
	State     state `json:"state"`
	Countdown int   `json:"countdown"`
}

// PayloadPlayerBet 廣播玩家的下注活動。
type PayloadPlayerBet struct {
	PlayerID  string          `json:"playerId"`
	BetAmount decimal.Decimal `json:"betAmount"`
	TotalBet  decimal.Decimal `json:"totalBet"`
}

// PayloadBetResult 是伺服器回傳給下注玩家的個人結果。
type PayloadBetResult struct {
	Success  bool            `json:"success"`
	Error    string          `json:"error,omitempty"`
	TotalBet decimal.Decimal `json:"totalBet,omitempty"`
	Balance  decimal.Decimal `json:"balance,omitempty"`
}

// PayloadOpening 廣播開獎結果。
type PayloadOpening struct {
	Number int `json:"number"`
}

// PayloadWinResult 廣播給贏家的中獎訊息。
type PayloadWinResult struct {
	BetAmount decimal.Decimal `json:"betAmount"`
	WinAmount decimal.Decimal `json:"winAmount"`
	Balance   decimal.Decimal `json:"balance"`
}
