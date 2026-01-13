package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type AdapterMode string

const (
	ModeMock AdapterMode = "mock"
	ModeReal AdapterMode = "real"
)

// AuthConfig 包含驗證服務的設定。
type AuthConfig struct {
	Mode AdapterMode `mapstructure:"mode"`
}

// DatabaseConfig 包含資料庫的設定。
type DatabaseConfig struct {
	Driver string `mapstructure:"driver"`
	DSN    string `mapstructure:"dsn"`
}

// ExternalWalletConfig 包含外接錢包 API 的設定。
type ExternalWalletConfig struct {
	BaseURL string `mapstructure:"baseUrl"`
	APIKey  string `mapstructure:"apiKey"`
}

// ExternalConfig 包含所有外部服務的設定。
type ExternalConfig struct {
	Wallet ExternalWalletConfig `mapstructure:"wallet"`
}

// RedisConfig 包含 Redis 的設定。
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// AppConfig 包含應用程式的所有全域設定。
//
// 這是一個聚合設定結構，包含了 WebSocket、資料庫、Redis 與外部服務等所有必要的設定。
// 它支援從設定檔 (yaml) 或環境變數中載入。
type AppConfig struct {
	// Port 是 WebSocket 服務監聽的連接埠。
	Port int `mapstructure:"port"`

	// WriteWaitSec 是 WebSocket 寫入逾時時間（秒）。
	WriteWaitSec int `mapstructure:"writeWaitSec"`

	// PongWaitSec 是 WebSocket Pong 回應等待時間（秒）。
	PongWaitSec int `mapstructure:"pongWaitSec"`

	// MaxMessageSize 是允許的最大 WebSocket 訊息大小（位元組）。
	MaxMessageSize int64 `mapstructure:"maxMessageSize"`

	// ReadBufferSize 是讀取緩衝區大小（位元組）。
	ReadBufferSize int `mapstructure:"readBufferSize"`

	// WriteBufferSize 是寫入緩衝區大小（位元組）。
	WriteBufferSize int `mapstructure:"writeBufferSize"`

	// Auth 包含驗證相關設定。
	Auth AuthConfig `mapstructure:"auth"`

	// Database 包含資料庫連線設定。
	Database DatabaseConfig `mapstructure:"database"`

	// External 包含外部 API 整合設定。
	External ExternalConfig `mapstructure:"external"`

	// Redis 包含 Redis 快取與 Pub/Sub 設定。
	Redis RedisConfig `mapstructure:"redis"`
}

// APIConfig 包含 REST API 伺服器的設定。
type APIConfig struct {
	Port int `mapstructure:"port"`
}

// LoadConfig 從指定路徑載入設定檔。
//
// 此函式使用泛型設計，可以載入任意結構的設定檔。
// 它會自動搜尋 config.{env}.yaml 並讀取環境變數覆寫設定。
//
// 參數說明：
//   - configPath: string, 設定檔所在的目錄路徑（不含檔名）。
//   - env: string, 環境名稱 (例如 "local", "dev", "prod")，決定讀取哪個 yaml 檔。
//
// 回傳值：
//   - *T: 成功載入的設定物件指標。
//   - error: 如果讀取或解析失敗，則返回錯誤。
func LoadConfig[T any](configPath string, env string) (*T, error) {
	v := viper.New() // 建議每次都 new 一個，避免 cmd 間衝突
	v.AddConfigPath(configPath)

	// 設定讀取的檔案名稱: config.{env}
	// 例如 env="local" -> config.local.yaml
	configName := fmt.Sprintf("config.%s", env)
	v.SetConfigName(configName)
	v.SetConfigType("yaml")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("無法讀取設定檔: %w", err)
	}

	var config T
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("無法解析設定檔: %w", err)
	}

	return &config, nil
}
