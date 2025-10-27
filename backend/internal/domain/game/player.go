package game

import (
	"encoding/json"

	"github.com/joe_shih/slot-factory/pkg/wss"
	"github.com/shopspring/decimal"
)

// Envelope 是所有 WebSocket 訊息的通用外層結構。
type Envelope struct {
	Action  string `json:"action"`
	Payload any    `json:"payload,omitempty"`
}

// Player 代表一個玩家實體。
type Player struct {
	// ID 是玩家的唯一識別碼。
	ID string
	// Name 是玩家的名稱。
	Name string
	// Point 是玩家餘額
	Point decimal.Decimal
	// Socket連線
	wss.Client
}

func (p *Player) SendMessage(message Envelope) error {
	data, _ := json.MarshalIndent(message, "", "  ")
	return p.Client.SendMessage(string(data))
}

// NewPlayer 創建一個新的 Player 實例。
//
// Params:
//   - id: string, 玩家的唯一識別碼。
//   - name: string, 玩家的名稱。
//
// Returns:
//   - *Player: 指向新建立的 Player 實例的指標。
func NewPlayer(id string, name string, client wss.Client) *Player {
	return &Player{
		ID:     id,
		Name:   name,
		Client: client,
	}
}
