package game

import (
	"encoding/json"
)

// Envelope 是所有 WebSocket 訊息的通用外層結構。
type Envelope struct {
	Action  string `json:"action"`
	Payload any    `json:"payload,omitempty"`
}

// GameClient 定義了與客戶端通訊所需實現的介面。
// 這將核心領域 (domain) 與具體的網路實現 (如 WebSocket) 分離。
type GameClient interface {
	SendMessage(message string) error
	Kick(reason string) error
	GetTag(key string) (value any, exists bool)
	SetTag(key string, value any)
	GetIP() string
}

// Player 代表一個玩家實體。
type Player struct {
	// ID 是玩家的唯一識別碼。
	ID string
	// Name 是玩家的名稱。
	Name string
	// client 是指向實現了 GameClient 介面的連線物件。
	client GameClient
}

// SendMessage 將一個 Envelope 結構序列化為 JSON 字串，並透過 GameClient 發送出去。
func (p *Player) SendMessage(message Envelope) error {
	data, _ := json.MarshalIndent(message, "", "  ")
	return p.client.SendMessage(string(data))
}

// Kick 透過 GameClient 踢出玩家連線。
func (p *Player) Kick(reason string) error {
	return p.client.Kick(reason)
}

// SetTag 透過 GameClient 為連線設置一個標籤。
func (p *Player) SetTag(key string, value any) {
	p.client.SetTag(key, value)
}

// GetTag 透過 GameClient 從連線讀取一個標籤。
func (p *Player) GetTag(key string) (value any, exists bool) {
	return p.client.GetTag(key)
}

// IP 透過 GameClient 從連線讀取IP
func (p *Player) IP() string {
	return p.client.GetIP()
}

// NewPlayer 創建一個新的 Player 實例。
func NewPlayer(id string, name string, client GameClient) *Player {
	return &Player{
		ID:     id,
		Name:   name,
		client: client,
	}
}
