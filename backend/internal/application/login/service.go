package login

import (
	"github.com/joe_shih/slot-factory/internal/domain/game"
	"github.com/joe_shih/slot-factory/pkg/wss"
)

// AuthClient 定義了外部身份驗證服務需要實現的介面。
// Use Case 會依賴此介面，而不是一個具體的驗證服務實作。
type AuthClient interface {
	// VerifyToken 驗證一個 token 並回傳使用者資料。
	//
	// Params:
	//   - token: string, 從客戶端傳來的驗證權杖。
	//
	// Returns:
	//   - UserData: 包含使用者 ID 和名稱的資料結構。
	//   - error: 如果驗證失敗，則回傳錯誤。
	VerifyToken(token string) (UserData, error)
}

// UserData 是從身份驗證服務成功回傳的使用者資訊。
type UserData struct {
	// ID 是使用者的唯一識別碼。
	ID string
	// Name 是使用者的名稱。
	Name string
}

// Service 提供了身份驗證相關的 use case。
type Service struct {
	authClient AuthClient
}

// NewService 創建一個新的 Service 實例。
//
// Params:
//   - client: AuthClient, 一個實現了 AuthClient 介面的外部服務客戶端。
//
// Returns:
//   - *Service: 新的 Service 實例。
func NewService(client AuthClient) *Service {
	return &Service{authClient: client}
}

// Authenticate 根據 token 驗證使用者身份，並回傳一個 domain 層的 Player 物件。
//
// Params:
//   - token: string, 要驗證的權杖。
//
// Returns:
//   - *game.Player: 代表已驗證玩家的 domain 物件。
//   - error: 如果驗證失敗，則回傳錯誤。
func (s *Service) Authenticate(token string, client wss.Client) (*game.Player, error) {
	data, err := s.authClient.VerifyToken(token)
	if err != nil {
		return nil, err
	}
	player := game.NewPlayer(data.ID, data.Name, client)
	return player, nil
}
