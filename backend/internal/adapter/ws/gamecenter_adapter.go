package ws

import (
	"github.com/joe_shih/slot-factory/internal/application/gamecenter"
	"github.com/joe_shih/slot-factory/pkg/wss"
)

// GameCenterAdapter 將來自 wss 層的事件，轉接給 application 層的 gamecenter.EventHandler。
// 它實現了 wss.Subscriber 介面，是標準的框架轉接器。
type GameCenterAdapter struct {
	handler gamecenter.EventHandler
}

// 確保 GameCenterAdapter 在編譯時期就實現了 wss.Subscriber 介面。
var _ wss.Subscriber = (*GameCenterAdapter)(nil)

// NewGameCenterAdapter 創建一個新的 GameCenterAdapter 實例。
func NewGameCenterAdapter(handler gamecenter.EventHandler) *GameCenterAdapter {
	return &GameCenterAdapter{handler: handler}
}

// OnConnect 在收到 wss 的連線事件時被呼叫。
func (a *GameCenterAdapter) OnConnect(client wss.Client) {
	// 將 wss.Client 轉接成 game.GameClient，然後再傳遞給核心應用層。
	gameClient := NewGameClientAdapter(client)
	a.handler.HandleConnect(gameClient)
}

// OnDisconnect 在收到 wss 的斷線事件時被呼叫。
func (a *GameCenterAdapter) OnDisconnect(client wss.Client) {
	gameClient := NewGameClientAdapter(client)
	a.handler.HandleDisconnect(gameClient)
}

// OnMessage 在收到 wss 的訊息事件時被呼叫。
func (a *GameCenterAdapter) OnMessage(client wss.Client, message []byte) {
	gameClient := NewGameClientAdapter(client)
	a.handler.HandleMessage(gameClient, message)
}
