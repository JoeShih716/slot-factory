package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http" // 為了 net/http 範例而導入
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joe_shih/slot-factory/pkg/config"
	"github.com/joe_shih/slot-factory/pkg/wss"
)



func main() {
	// 1. 初始化結構化日誌 Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// 2. 載入設定檔
	cfg, err := config.LoadConfig("./configs")
	if err != nil {
		logger.Error("cannot load config", "error", err)
		os.Exit(1)
	}
	port := cfg.Websocket.Port

	// 3. 建立一個 context 用於控制伺服器生命週期
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // 確保在 main 函式結束時 context 被取消

	// 4. 建立 WebSocket 伺服器設定
	wssConfig := &wss.Config{
		WriteWait:       time.Duration(cfg.Websocket.WriteWaitSec) * time.Second,
		PongWait:        time.Duration(cfg.Websocket.PongWaitSec) * time.Second,
		MaxMessageSize:  cfg.Websocket.MaxMessageSize,
		ReadBufferSize:  cfg.Websocket.ReadBufferSize,
		WriteBufferSize: cfg.Websocket.WriteBufferSize,
	}

	// 5. 建立 WebSocket 伺服器實例
	wsServer := wss.NewServer(ctx, wssConfig, logger)

	// 6. 建立並註冊業務邏輯處理器
	gameHandler := &dummyGameHandler{logger: logger}
	wsServer.RegisterHandler(gameHandler)

	// --- 方法一：使用 Gin 框架 (啟用中) ---
	runWithGin(logger, wsServer, port)

	// --- 方法二：使用標準庫 net/http (已註解) ---
	// runWithNetHTTP(logger, wsServer, port)
}

// runWithGin 使用 Gin 框架來啟動伺服器。
func runWithGin(logger *slog.Logger, handler http.Handler, port int) {
	logger.Info("starting server with gin framework")
	engine := gin.Default()

	// 使用 gin.WrapH 將實現了 http.Handler 的 wsServer 包裝成 Gin 的 HandlerFunc
	engine.GET("/ws", gin.WrapH(handler))

	logger.Info("websocket server (gin) starting", "port", port)
	if err := engine.Run(fmt.Sprintf(":%d", port)); err != nil {
		logger.Error("failed to run gin server", "error", err)
		os.Exit(1)
	}
}

// runWithNetHTTP 使用標準庫 net/http 來啟動伺服器。
func runWithNetHTTP(logger *slog.Logger, handler http.Handler, port int) {
	logger.Info("starting server with net/http framework")
	mux := http.NewServeMux()

	// 直接將 wsServer 作為一個 http.Handler 註冊到路由
	mux.Handle("/ws", handler)

	logger.Info("websocket server (net/http) starting", "port", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		logger.Error("failed to run net/http server", "error", err)
		os.Exit(1)
	}
}
