package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config 應用程式的整體設定結構。
type Config struct {
	Websocket WebsocketConfig `mapstructure:"websocket"`
	API       APIConfig       `mapstructure:"api"`
	// 未來可以加入資料庫、Redis 等設定
}

// WebsocketConfig 包含 WebSocket 伺服器的設定。
type WebsocketConfig struct {
	Port            int   `mapstructure:"port"`
	WriteWaitSec    int   `mapstructure:"writeWaitSec"`
	PongWaitSec     int   `mapstructure:"pongWaitSec"`
	MaxMessageSize  int64 `mapstructure:"maxMessageSize"`
	ReadBufferSize  int   `mapstructure:"readBufferSize"`
	WriteBufferSize int   `mapstructure:"writeBufferSize"`
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
func LoadConfig(configPath string) (config *Config, err error) {
	viper.AddConfigPath(configPath)
	viper.SetConfigName("config") // 設定檔名稱為 config.yaml 或 config.json
	viper.SetConfigType("yaml")   // 預設使用 YAML 格式

	viper.AutomaticEnv() // 自動讀取環境變數

	err = viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("無法讀取設定檔: %w", err)
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("無法解析設定檔: %w", err)
	}

	return config, nil
}
