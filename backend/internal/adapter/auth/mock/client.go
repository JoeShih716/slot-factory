package mock

import (
	"strconv"

	"github.com/joe_shih/slot-factory/internal/application/login"
)

// AuthClient 是一個 login.AuthClient 的模擬實作，用於測試和本地開發。
type AuthClient struct {
	// 流水號
	counterID int
}

// NewAuthClient 創建一個新的 AuthClient 實例。
func NewAuthClient() *AuthClient {
	return &AuthClient{counterID: 1000000}
}

// VerifyToken 模擬驗證 token 的過程，並始終回傳一個固定的假使用者資料。
func (c *AuthClient) VerifyToken(token string) (login.UserData, error) {
	// 在模擬版本中，我們忽略 token，直接回傳成功
	c.counterID++
	strID := strconv.Itoa(c.counterID)
	userData := login.UserData{
		ID:   strID,
		Name: "MockPlayer" + strID,
	}
	return userData, nil
}
