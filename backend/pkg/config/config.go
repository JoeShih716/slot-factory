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

// WebsocketConfig 包含 WebSocket 伺服器的設定。
type WebsocketConfig struct {
	Port            int         `mapstructure:"port"`
	WriteWaitSec    int         `mapstructure:"writeWaitSec"`
	PongWaitSec     int         `mapstructure:"pongWaitSec"`
	MaxMessageSize  int64       `mapstructure:"maxMessageSize"`
	ReadBufferSize  int         `mapstructure:"readBufferSize"`
	WriteBufferSize int         `mapstructure:"writeBufferSize"`
	Mode            AdapterMode `mapstructure:"mode"`
}

// APIConfig 包含 API 伺服器的設定。
type APIConfig struct {
	Port int `mapstructure:"port"`
}

// LoadConfig 從指定路徑載入設定檔。
//
// 參數說明：
//   - configPath: string, 設定檔所在的目錄路徑。
//
// 回傳值：
//   - *Config: 載入的設定物件。
//   - error: 如果載入失敗，則返回錯誤。
//
// LoadConfig 使用泛型回傳指定類型的設定。
func LoadConfig[T any](configPath string) (*T, error) {
	v := viper.New() // 建議每次都 new 一個，避免 cmd 間衝突
	v.AddConfigPath(configPath)
	v.SetConfigName("config") // 例如 config.yaml
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
