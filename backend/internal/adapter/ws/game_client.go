package ws

import (
	"net"
	"strings"

	"github.com/joe_shih/slot-factory/internal/domain/game"
	"github.com/joe_shih/slot-factory/pkg/wss"
)

// GameClientAdapter 將一個 wss.Client 物件「轉接」成一個 game.GameClient。
// 它扮演了轉接頭的角色，填補了兩個介面之間的差異。
type GameClientAdapter struct {
	client wss.Client
}

// 確保 GameClientAdapter 在編譯時期就實現了 game.GameClient 介面。
var _ game.GameClient = (*GameClientAdapter)(nil)

// NewGameClientAdapter 創建一個新的轉接器實例。
func NewGameClientAdapter(client wss.Client) game.GameClient {
	return &GameClientAdapter{client: client}
}

// --- 實現 game.GameClient 介面 ---

// SendMessage 直接呼叫底層 client 的同名方法。
func (a *GameClientAdapter) SendMessage(message string) error {
	return a.client.SendMessage(message)
}

// Kick 直接呼叫底層 client 的同名方法。
func (a *GameClientAdapter) Kick(reason string) error {
	return a.client.Kick(reason)
}

// SetTag 直接呼叫底層 client 的同名方法。
func (a *GameClientAdapter) SetTag(key string, value any) {
	a.client.SetTag(key, value)
}

// GetTag 直接呼叫底層 client 的同名方法。
func (a *GameClientAdapter) GetTag(key string) (any, bool) {
	return a.client.GetTag(key)
}

// GetIP 是此 Adapter 的核心轉接邏輯。
// 它呼叫底層 client 的 RemoteAddr() 方法，並從中解析出 IP 位址。
func (a *GameClientAdapter) GetIP() string {
	addr := a.client.RemoteAddr()
	// net.SplitHostPort 對 IPv6 的位址 (例如 "[::1]:1234") 也能正常處理
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		// 如果解析失敗 (例如 addr 不含 port)，則嘗試直接回傳 addr，
		// 因為它可能本身就是一個 IP 位址。
		// 移除 IPv6 的方括號 (如果有的話)
		return strings.Trim(addr, "[]")
	}
	return host
}
