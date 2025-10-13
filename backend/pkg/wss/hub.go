package wss

import (
	"context"
	"log/slog"
)

// hub 維護一組活躍的客戶端，並將事件分派給所有已註冊的 Handler。
type hub struct {
	clients    map[*connection]bool
	register   chan *connection
	unregister chan *connection
	inbound    chan *clientMessage
	handlers   []Handler
	ctx        context.Context
	logger     *slog.Logger
}

// newHub 創建一個新的 hub 實例。
func newHub(ctx context.Context, logger *slog.Logger) *hub {
	return &hub{
		register:   make(chan *connection),
		unregister: make(chan *connection),
		inbound:    make(chan *clientMessage),
		clients:    make(map[*connection]bool),
		handlers:   make([]Handler, 0),
		ctx:        ctx,
		logger:     logger,
	}
}

// registerHandler 註冊一個新的事件處理器。
func (h *hub) registerHandler(handler Handler) {
	if handler != nil {
		h.handlers = append(h.handlers, handler)
	}
}

// run 啟動 hub 的事件處理迴圈。
func (h *hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.logger.Info("client registered", "clientID", client.ID())
			for _, handler := range h.handlers {
				handler.OnConnect(client)
			}
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				h.logger.Info("client unregistered", "clientID", client.ID())
				for _, handler := range h.handlers {
					handler.OnDisconnect(client)
				}
			}
		case msg := <-h.inbound:
			h.logger.Debug("message received from client", "clientID", msg.client.ID())
			for _, handler := range h.handlers {
				handler.OnMessage(msg.client, msg.message)
			}
		case <-h.ctx.Done():
			// Context 被取消，開始關閉程序
			h.logger.Info("hub shutting down")
			for client := range h.clients {
				client.Kick("Server is shutting down.")
				delete(h.clients, client)
				close(client.send)
			}
			return // 結束 run 迴圈
		}
	}
}
