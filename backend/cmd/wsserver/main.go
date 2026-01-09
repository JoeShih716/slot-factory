package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joe_shih/slot-factory/internal/adapter/auth/mock"
	"github.com/joe_shih/slot-factory/internal/adapter/auth/real"
	walletMock "github.com/joe_shih/slot-factory/internal/adapter/wallet/mock"
	"github.com/joe_shih/slot-factory/internal/adapter/ws"
	"github.com/joe_shih/slot-factory/internal/application/gamecenter"
	"github.com/joe_shih/slot-factory/internal/application/login"
	"github.com/joe_shih/slot-factory/internal/application/wallet"
	"github.com/joe_shih/slot-factory/internal/gameImp/game1000"
	"github.com/joe_shih/slot-factory/internal/gameImp/game1001"
	"github.com/joe_shih/slot-factory/pkg/config"
	"github.com/joe_shih/slot-factory/pkg/wss"
)

// --- Main Application Setup ---

const configPath = "./configs/wsServer"

func main() {
	// 1. 初始化結構化日誌 Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// 2. 決定當前環境 (default: local)
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local"
	}
	logger.Info("loading config", "env", env)

	// 3. 載入設定檔 (e.g., config.local.yaml)
	cfg, err := config.LoadConfig[config.WebsocketConfig](configPath, env)
	if err != nil {
		logger.Error("cannot load config", "error", err)
		os.Exit(1)
	}
	port := cfg.Port

	// 業界慣例：優先讀取環境變數 PORT (常見於 Cloud Run, Heroku, K8s)
	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
			logger.Info("using port from environment variable", "port", port)
		} else {
			logger.Warn("invalid PORT environment variable, using config value", "envPort", envPort, "configPort", port)
		}
	}

	// 3. 建立一個 context 用於監聽中斷信號，以實現優雅關機
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 4. 根據設定檔初始化底層 Adapters
	var authClient login.AuthClient
	var payment wallet.Payment

	mode := cfg.Mode
	if envMode := os.Getenv("MODE"); envMode != "" {
		logger.Info("using mode from environment variable", "mode", envMode)
		mode = config.AdapterMode(envMode)
	}
	switch mode {
	case config.ModeReal:
		authClient = real.NewAuthClient()
		logger.Info("using REAL auth adapter")
	default:
		authClient = mock.NewAuthClient()
		payment = walletMock.NewPayment()
		logger.Info("using MOCK auth adapter")
	}

	// 5. 建立 Application Services (核心業務邏輯)
	loginService := login.NewService(authClient)
	walletService := wallet.NewService(logger, payment)
	gameCenterService := gamecenter.NewService(*loginService, logger.With("component", "game_center"))

	// 6. 註冊所有遊戲實例到 Game Center
	gameCenterService.RegisterGame(game1000.NewGame(logger, walletService))
	gameCenterService.RegisterGame(game1001.NewGame(logger, walletService))

	// 7. 建立 WebSocket 伺服器
	wssConfig := &wss.Config{
		WriteWait:       time.Duration(cfg.WriteWaitSec) * time.Second,
		PongWait:        time.Duration(cfg.PongWaitSec) * time.Second,
		MaxMessageSize:  cfg.MaxMessageSize,
		ReadBufferSize:  cfg.ReadBufferSize,
		WriteBufferSize: cfg.WriteBufferSize,
	}
	wsServer := wss.NewServer(ctx, wssConfig, logger.With("component", "wss"))

	// 8. 建立框架轉接器，並將其註冊到 WebSocket 伺服器
	wsAdapter := ws.NewGameCenterAdapter(gameCenterService)
	wsServer.Register(wsAdapter)

	// 9. 設定 Gin 引擎並掛載 WebSocket Handler
	engine := gin.Default()
	engine.GET("/ws", gin.WrapH(wsServer))

	// 10. 建立並啟動 HTTP 伺服器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: engine,
	}

	go func() {
		logger.Info("http server starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("failed to run http server", "error", err)
			os.Exit(1)
		}
	}()

	// 11. 等待中斷信號，執行優雅關機
	<-ctx.Done()
	stop()
	logger.Info("shutting down gracefully, press Ctrl+C again to force")

	// 設定一個超時 context
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown failed", "error", err)
	}

	logger.Info("server exiting")
}
