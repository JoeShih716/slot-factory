package gamecenter

import "github.com/joe_shih/slot-factory/internal/domain/game"

// EventHandler 定義了 gamecenter 處理外部連線事件所需實現的介面。
// 這是 application 層的入口點 (port)，由外部的 adapter 來驅動。
type EventHandler interface {
	HandleConnect(client game.GameClient)
	HandleDisconnect(client game.GameClient)
	HandleMessage(client game.GameClient, message []byte)
}
