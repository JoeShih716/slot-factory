package wss

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// Config 定義了 WebSocket 伺服器的所有可設定參數。
type Config struct {
	WriteWait       time.Duration // 寫入操作的超時時間
	PongWait        time.Duration // 等待 Pong 訊息的超時時間
	PingPeriod      time.Duration // 發送 Ping 訊息的間隔
	MaxMessageSize  int64         // 允許接收的最大訊息大小
	ReadBufferSize  int           // 讀取緩衝區的大小
	WriteBufferSize int           // 寫入緩衝區的大小
}

// Server 是 websocket package 對外的主要門面 (Facade)，並實現了 http.Handler 介面。
type Server struct {
	hub    *hub
	cfg    *Config
	logger *slog.Logger
}

// NewServer 創建並設定一個完整的 WebSocket 伺服器。
func NewServer(ctx context.Context, cfg *Config, logger *slog.Logger) *Server {
	// 如果 PingPeriod 沒有被設定，則根據 PongWait 計算一個合理的值
	if cfg.PingPeriod == 0 && cfg.PongWait > 0 {
		cfg.PingPeriod = (cfg.PongWait * 9) / 10
	}

	h := newHub(ctx, logger.With("component", "hub"))
	go h.run()

	return &Server{
		hub:    h,
		cfg:    cfg,
		logger: logger.With("component", "wss_server"),
	}
}

// RegisterHandler 將一個業務邏輯處理器註冊到 WebSocket 伺服器。
func (s *Server) RegisterHandler(handler Handler) {
	s.hub.registerHandler(handler)
}

// ServeHTTP 實現 http.Handler 介面，處理 WebSocket 的升級請求。
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  s.cfg.ReadBufferSize,
		WriteBufferSize: s.cfg.WriteBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			// 在生產環境中，這裡應該有更嚴格的來源檢查
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Warn("websocket upgrade failed", "error", err)
		return
	}

	clientLogger := s.logger.With("component", "client")
	client := newConnection(s.hub, conn, r, clientLogger)
	client.hub.register <- client

	go client.writePump(s.cfg)
	go client.readPump(s.cfg)
}
