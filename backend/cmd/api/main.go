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

	// 載入設定 (API	// 2. 載入共用設定 (AppConfig)
	appCfg, err := config.LoadConfig[config.AppConfig](configPath, env)
	if err != nil {
		logger.Error("failed to load app config", "error", err)
		os.Exit(1)
	}

	// API 預設 Port 8081 (避免與 wsserver 8080 衝突)
	port := 8081
	envPort := os.Getenv("API_PORT")
	if envPort != "" {
		port, _ = strconv.Atoi(envPort)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 初始化資料庫
	var db *gorm.DB
	if appCfg.Database.Driver == "mysql" || appCfg.Database.Driver == "proxy" {
		d, err := gorm.Open(mysql.Open(appCfg.Database.DSN), &gorm.Config{})
		if err != nil {
			logger.Error("failed to connect to mysql", "error", err)
			os.Exit(1)
		}
		db = d
	}

	// Redis (如果設定檔有填寫)
	var rdb *redis.Client
	if appCfg.Redis.Addr != "" {
		rdb = redis.NewClient(&redis.Options{
			Addr:     appCfg.Redis.Addr,
			Password: appCfg.Redis.Password,
			DB:       appCfg.Redis.DB,
		})
		logger.Info("connected to redis", "addr", appCfg.Redis.Addr)
	}

	// Auth (Mock) - API 服務這裡暫時用 Mock，實際上可能需要驗證管理員 Token
	// 如果需要 Real Auth，可從 appCfg.Auth.Mode 判斷
	authClient := authMock.NewAuthClient()

	// External Wallet (Proxy)
	// 如果 DB Driver 是 proxy，則需要初始化 ProxyPayment
	var payment wallet.Payment
	if appCfg.Database.Driver == "proxy" {
		payment = walletProxy.NewPayment(db, appCfg.External.Wallet.BaseURL, appCfg.External.Wallet.APIKey)
		logger.Info("using PROXY (External API + Local Log) adapter")
	} else {
		payment = walletMock.NewPayment()
	}

	// 初始化 Services
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
	err = srv.Shutdown(shutdownCtx)
	if err != nil {
		logger.Error("API server shutdown failed", "error", err)
		os.Exit(1)
	}
}
