package wss

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// connection 是 Client 介面的具體實現，負責管理底層 WebSocket 連線。
type connection struct {
	id         string
	hub        *hub
	conn       *websocket.Conn
	send       chan []byte
	mu         sync.Mutex
	remoteAddr string
	headers    http.Header
	tags       map[string]any
	tagsMutex  sync.RWMutex
	logger     *slog.Logger
}

// 確保 connection 類型在編譯時期就實現了 Client 接口。
var _ Client = (*connection)(nil)

// newConnection 創建一個新的客戶端連線實例。
func newConnection(hub *hub, conn *websocket.Conn, r *http.Request, logger *slog.Logger) *connection {
	clientID := generateClientID()
	return &connection{
		id:         clientID,
		hub:        hub,
		conn:       conn,
		send:       make(chan []byte, 256),
		remoteAddr: r.RemoteAddr,
		headers:    r.Header.Clone(), // 複製標頭以確保安全
		tags:       make(map[string]any),
		logger:     logger.With("clientID", clientID),
	}
}

// ID 返回客戶端的唯一標識符。
func (c *connection) ID() string {
	return c.id
}

// SendMessage 將一則訊息放入發送佇列，由 writePump 異步發送。
func (c *connection) SendMessage(message string) error {
	c.send <- []byte(message)
	return nil
}

// Kick 立即中斷與客戶端的連線。
func (c *connection) Kick(reason string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, reason))
}

// RemoteAddr 返回客戶端的網路位址。
func (c *connection) RemoteAddr() string {
	return c.remoteAddr
}

// Headers 返回客戶端升級請求時的 HTTP 標頭。
func (c *connection) Headers() http.Header {
	return c.headers
}

// UserAgent 返回客戶端的 User-Agent。
func (c *connection) UserAgent() string {
	return c.headers.Get("User-Agent")
}

// SetTag 在該連線的生命週期內附加一個鍵值對資料。
func (c *connection) SetTag(key string, value any) {
	c.tagsMutex.Lock()
	defer c.tagsMutex.Unlock()
	c.tags[key] = value
}

// GetTag 根據鍵名讀取之前用 SetTag 附加的資料。
func (c *connection) GetTag(key string) (value any, exists bool) {
	c.tagsMutex.RLock()
	defer c.tagsMutex.RUnlock()
	value, exists = c.tags[key]
	return
}

// readPump 將來自 WebSocket 連線的訊息泵送到 hub。
func (c *connection) readPump(cfg *Config) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(cfg.MaxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(cfg.PongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(cfg.PongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Warn("read pump failed", "error", err)
			}
			break
		}
		c.hub.inbound <- &clientMessage{client: c, message: message}
	}
}

// writePump 將來自 hub 的訊息泵送到 WebSocket 連線。
func (c *connection) writePump(cfg *Config) {
	ticker := time.NewTicker(cfg.PingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.mu.Lock()
			c.conn.SetWriteDeadline(time.Now().Add(cfg.WriteWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				c.mu.Unlock()
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				c.logger.Warn("write pump failed on getting next writer", "error", err)
				c.mu.Unlock()
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				c.logger.Warn("write pump failed on closing writer", "error", err)
				c.mu.Unlock()
				return
			}
			c.mu.Unlock()
		case <-ticker.C:
			c.mu.Lock()
			c.conn.SetWriteDeadline(time.Now().Add(cfg.WriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.logger.Warn("write pump failed on sending ping", "error", err)
				c.mu.Unlock()
				return
			}
			c.mu.Unlock()
		}
	}
}

// generateClientID 創建一個唯一的、基於 UUID 的客戶端 ID。
func generateClientID() string {
	return uuid.NewString()
}
