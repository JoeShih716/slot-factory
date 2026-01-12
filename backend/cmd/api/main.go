package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	authMock "github.com/joe_shih/slot-factory/internal/adapter/auth/mock"
	internalHTTP "github.com/joe_shih/slot-factory/internal/adapter/http"
	walletMock "github.com/joe_shih/slot-factory/internal/adapter/wallet/mock"
	walletProxy "github.com/joe_shih/slot-factory/internal/adapter/wallet/proxy"
	"github.com/joe_shih/slot-factory/internal/application/gamecenter"
	"github.com/joe_shih/slot-factory/internal/application/login"
	"github.com/joe_shih/slot-factory/internal/application/wallet"
	"github.com/joe_shih/slot-factory/internal/config"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const configPath = "./configs"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local"
	}

	// 載入設定 (API 服務同樣使用 WebsocketConfig 中的資料庫與 Redis 設定)
	cfg, err := config.LoadConfig[config.WebsocketConfig](configPath, env)
	if err != nil {
		logger.Error("cannot load config", "error", err)
		os.Exit(1)
	}

	// API 預設 Port 8081 (避免與 wsserver 8080 衝突)
	port := 8081
	if envPort := os.Getenv("API_PORT"); envPort != "" {
		fmt.Sscanf(envPort, "%d", &port)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 初始化資料庫
	var db *gorm.DB
	if cfg.Database.Driver == "mysql" || cfg.Database.Driver == "proxy" {
		d, err := gorm.Open(mysql.Open(cfg.Database.DSN), &gorm.Config{})
		if err != nil {
			logger.Error("failed to connect to mysql", "error", err)
			os.Exit(1)
		}
		db = d
	}

	// 初始化 Redis
	var rdb *redis.Client
	if cfg.Redis.Addr != "" {
		rdb = redis.NewClient(&redis.Options{
			Addr:     cfg.Redis.Addr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})
	}

	// 初始化 Adapters (API 服務通常只需要 Read-only 或特定介面)
	var payment wallet.Payment
	if cfg.Database.Driver == "proxy" {
		payment = walletProxy.NewPayment(db, cfg.External.Wallet.BaseURL, cfg.External.Wallet.APIKey)
	} else {
		payment = walletMock.NewPayment()
	}

	// 初始化 Services
	authClient := authMock.NewAuthClient() // API 服務暫時用 Mock
	loginService := login.NewService(authClient)
	walletService := wallet.NewService(logger, payment)
	gameCenterService := gamecenter.NewService(*loginService, logger.With("component", "game_center"), rdb)

	// 設定 Gin
	engine := gin.Default()
	handler := internalHTTP.NewHandler(gameCenterService, gameCenterService, walletService)

	apiV1 := engine.Group("/api/v1")
	{
		apiV1.GET("/games", handler.HandleGetGames)
		apiV1.GET("/history", handler.HandleGetHistory)
		apiV1.POST("/admin/kick_all", handler.HandleKickAll)
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: engine,
	}

	go func() {
		logger.Info("Independent API server starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("API server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("Shutting down API server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}
