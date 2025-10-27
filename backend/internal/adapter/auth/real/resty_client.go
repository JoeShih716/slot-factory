package real

import (
	"github.com/go-resty/resty/v2"
	"github.com/joe_shih/slot-factory/internal/application/login"
)

// AuthClient 是一個使用 resty 的 login.AuthClient 實作。
type AuthClient struct {
	client *resty.Client
}

// NewAuthClient 創建一個新的 AuthClient 實例。
func NewAuthClient() *AuthClient {
	return &AuthClient{
		client: resty.New(),
	}
}

// VerifyToken 使用 resty client 驗證 token。
// TODO: 目前這個函式只回傳假資料，未來需要實現真正的 HTTP 請求邏輯。
func (c *AuthClient) VerifyToken(token string) (login.UserData, error) {
	// 在未來的實作中，這裡會發送一個 HTTP 請求到外部驗證服務
	// resp, err := c.client.R().
	// 	 SetAuthToken(token).
	// 	 SetResult(&login.UserData{}).
	// 	 Post("https://your-auth-service.com/verify")

	// 目前回傳一個固定的假資料
	userData := login.UserData{
		ID:   "real_user_456",
		Name: "Resty Client",
	}
	return userData, nil
}
