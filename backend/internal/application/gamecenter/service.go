package gamecenter

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/joe_shih/slot-factory/internal/application/login"
	"github.com/joe_shih/slot-factory/internal/domain/game"
	"github.com/joe_shih/slot-factory/pkg/wss"
	"github.com/shopspring/decimal"
)

// 確保 service 類型在編譯時期就實現了 wss.Subscriber 接口。
var _ wss.Subscriber = (*service)(nil)

// Service 定義了遊戲核心業務邏輯的介面。
type Service interface {
	// JoinGame 讓一個玩家加入指定的遊戲。
	//
	// Params:
	//   - ctx: context.Context, 請求上下文。
	//   - gameID: int, 玩家要加入的遊戲 ID。
	//   - player: domain.Player, 要加入遊戲的玩家。
	//
	// Returns:
	//   - error: 如果發生錯誤則回傳 error。
	JoinGame(ctx context.Context, gameID int, player game.Player) error

	// LeaveGame 讓一個玩家離開指定的遊戲。
	//
	// Params:
	//   - ctx: context.Context, 請求上下文。
	//   - gameID: int, 玩家要離開的遊戲 ID。
	//   - player: domain.Player, 要離開遊戲的玩家。
	//
	// Returns:
	//   - error: 如果發生錯誤則回傳 error。
	LeaveGame(ctx context.Context, gameID int, player game.Player) error
}

// service 是 Service 介面的具體實現。
type service struct {
	loginService login.Service
	logger       *slog.Logger
	games        map[int]game.IGame
}

// NewService 創建一個新的 game service 實例。
//
// Params:
//   - gameRepo: GameRepository, 遊戲儲存庫的介面實作。
//
// Returns:
//   - Service: 新的 game service 實例。
func NewService(loginService login.Service, logger *slog.Logger) *service {
	return &service{
		loginService: loginService,
		logger:       logger,
		games:        make(map[int]game.IGame),
	}
}

func (s *service) OnConnect(client wss.Client) {
	s.logger.Info("game service: client connected", "clientID", client.ID())
}

func (s *service) OnDisconnect(client wss.Client) {
	s.logger.Info("game service: client disconnected", "clientID", client.ID())
	// 在這裡可以加入玩家離線的處理邏輯，例如從遊戲中移除
	player, _ := client.GetTag("player")
	if player != nil {
		s.leaveGame(*(player.(*game.Player)))
	}
}

func (s *service) OnMessage(client wss.Client, message []byte) {
	// 先解析 action 層
	var base struct {
		Action ActionType      `json:"action"`
		Data   json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(message, &base); err != nil {
		s.logger.Warn("failed to unmarshal message", "error", err, "clientID", client.ID())
		client.Kick("invalid message format")
		return
	}

	s.logger.Info("message received", "action", base.Action, "clientID", client.ID())

	switch base.Action {
	case Login:
		var payload loginPayload
		if err := json.Unmarshal(base.Data, &payload); err != nil {
			s.logger.Warn("invalid auth payload", "error", err, "clientID", client.ID())
			client.SendMessage(`{"error": "invalid auth payload"}`)
			return
		}
		s.handleLogin(client, payload.Sid)
		player, _ := client.GetTag("player")
		if player != nil {
			domainPlayer := *(player.(*game.Player))
			s.joinGame(payload.GameID, domainPlayer)
		}
	case Play:
		var payload playPayload
		if err := json.Unmarshal(base.Data, &payload); err != nil {
			s.logger.Warn("invalid auth payload", "error", err, "clientID", client.ID())
			client.SendMessage(`{"error": "invalid auth payload"}`)
			return
		}
		s.handlePlay(client, payload.BetAmount)
	default:
		s.logger.Warn("unknown action", "action", base.Action, "clientID", client.ID())
		client.SendMessage("{\"error\": \"unknown action\"}")
	}
}

func (s *service) handleLogin(client wss.Client, token string) {
	if token == "" {
		client.Kick("auth failed: token is missing")
		return
	}

	player, err := s.loginService.Authenticate(token, client)
	if err != nil {
		s.logger.Error("authentication failed", "error", err, "clientID", client.ID())
		client.Kick("authentication failed")
		return
	}

	// 將驗證成功的 Player 物件附加到連線上
	client.SetTag("player", player)
	s.logger.Info("client authenticated successfully", "playerID", player.ID, "playerName", player.Name, "clientID", client.ID())

	response := fmt.Sprintf("{\"message\": \"authenticated successfully\", \"playerID\": \"%s\"}", player.ID)
	client.SendMessage(response)
}

func (s *service) handlePlay(client wss.Client, betAmount decimal.Decimal) {

	player, _ := client.GetTag("player")
	if player == nil {
		client.Kick("Not Login")
	}
	domainPlayer := (player.(*game.Player))
	gameID, exists := domainPlayer.GetTag("game")
	if !exists {
		client.Kick("Not Login")
	}
	realGameID := gameID.(int)
	game := s.games[realGameID]
	if game == nil {
		fmt.Println("handlePlay game not found ", gameID)
		domainPlayer.Kick("handlePlay game not found ")
		return
	}
	game.Play(domainPlayer, betAmount)
}

func (s service) RegisterGame(game game.IGame) {
	id := game.ID()
	s.games[id] = game
}

func (s *service) joinGame(gameID int, player game.Player) error {
	game := s.games[gameID]
	if game == nil {
		player.Kick("joinGame game not found")
		return fmt.Errorf("game not found")
	}
	game.AddPlayer(&player)
	player.SetTag("game", gameID)
	return nil
}

func (s *service) leaveGame(player game.Player) error {
	gameID, exists := player.GetTag("game")
	if !exists {
		return nil
	}
	realGameID := gameID.(int)
	game := s.games[realGameID]
	game.RemovePlayer(&player)
	return nil
}
