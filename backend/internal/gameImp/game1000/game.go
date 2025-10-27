package game1000

import (
	"math/rand/v2"

	"github.com/joe_shih/slot-factory/internal/domain/game"
	"github.com/shopspring/decimal"
)

const (
	// ActionPlayResult 是此遊戲回傳結果時使用的動作類型。
	ActionPlayResult = "play_result"
)

// Game 實作一個簡單的單人骰子遊戲 (game.IGame 介面)。
// 這是 game.IGame 的一個最簡實現，用於展示單人遊戲的邏輯。
type Game struct {
	id int
}

// NewGame 創建一個新的 1000 骰子遊戲實例。
func NewGame() game.IGame {
	return &Game{
		id: 1000,
	}
}

// ID 返回遊戲的唯一標識符。
func (g *Game) ID() int {
	return g.id
}

// AddPlayer 在單人遊戲中，此方法僅發送歡迎訊息，不需將玩家存儲在遊戲狀態中。
func (g *Game) AddPlayer(player *game.Player) {
	welcomeMsg := game.Envelope{
		Action:  "welcome",
		Payload: "welcome to dice game (Game 1000)",
	}
	player.SendMessage(welcomeMsg)
}

// RemovePlayer 在單人遊戲中，此方法為空，因為沒有需要從遊戲中清理的玩家狀態。
func (g *Game) RemovePlayer(player *game.Player) {
	// 單人遊戲，無共享狀態，不需實作
}

// playResult 是此遊戲的結果訊息結構。
type playResult struct {
	Success   bool            `json:"success"`
	Error     string          `json:"error,omitempty"`
	BetAmount decimal.Decimal `json:"betAmount"`
	WinAmount decimal.Decimal `json:"winAmount"`
	Dice      int             `json:"dice"`
}

// Play 處理玩家的遊玩請求，執行一次完整的單人遊戲流程。
// 遊戲邏輯：骰出一個 1~6 的數字，如果結果為 1，玩家贏得 6 倍賭注。
func (g *Game) Play(player *game.Player, betAmount decimal.Decimal) {
	var result playResult

	if betAmount.LessThanOrEqual(decimal.Zero) {
		result = playResult{
			Success: false,
			Error:   "Bet amount must be greater than zero.",
		}
	} else {
		// 執行遊戲核心邏輯
		dice := rand.IntN(6) + 1 // 產生1到6的隨機數
		winAmount := decimal.Zero
		if dice == 1 {
			winAmount = betAmount.Mul(decimal.NewFromInt(6))
		}
		result = playResult{
			Success:   true,
			BetAmount: betAmount,
			WinAmount: winAmount,
			Dice:      dice,
		}
	}

	// 將結果包裝在標準的 Envelope 中發送給客戶端
	player.SendMessage(game.Envelope{
		Action:  ActionPlayResult,
		Payload: result,
	})
}

// 確保 Game 類型在編譯時期就實現了 IGame 接口。
var _ game.IGame = (*Game)(nil)
