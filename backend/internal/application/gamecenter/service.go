package gamecenter

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/joe_shih/slot-factory/internal/application/login"
	"github.com/joe_shih/slot-factory/internal/domain/game"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
)

// Redis 相關常數
const (
	RedisKeyPlayerCountPrefix = "games:%d:count"
	RedisChannelControl       = "game_control"
)

// ControlCommand 代表從 Redis Pub/Sub 接收到的全域指令
type ControlCommand struct {
	Action string `json:"action"`
	Data   string `json:"data"`
}

// GameInfo 包含遊戲的基本資訊。
type GameInfo struct {
	ID          string `json:"id"`
	PlayerCount int    `json:"playerCount"`
}

// 確保 service 類型在編譯時期就實現了 EventHandler 接口。
var _ EventHandler = (*gameCenter)(nil)

// GameProvider 提供了讀取遊戲列表的介面。
type GameProvider interface {
	GetGames(ctx context.Context) []GameInfo
}

// AdminProvider 提供了系統管理相關的介面。
type AdminProvider interface {
	KickAll(ctx context.Context) error
}

// Service 定義了遊戲核心業務邏輯的介面（整合型，供 wsserver 使用）。
type Service interface {
	GameProvider
	AdminProvider
	// RegisterGame 註冊一個遊戲。
	RegisterGame(game game.IGame)
}

// gameCenter 是 Service 介面的具體實現。
type gameCenter struct {
	loginService login.Service
	logger       *slog.Logger
	redisClient  *redis.Client
	games        map[int]game.IGame
	clientList   map[string]game.GameClient
}

// NewService 建立並初始化一個新的遊戲中心服務實例。
//
// 此函式負責依賴注入，並在 Redis 客戶端可用時，啟動背景 goroutine 監聽全域控制指令。
//
// 參數說明：
//   - loginService: login.Service, 負責玩家登入驗證的服務。
//   - logger: *slog.Logger, 用於記錄日誌的 Logger 實例。
//   - rdb: *redis.Client, Redis 客戶端，用於全域計數與廣播。如果為 nil，則相關功能將被略過。
//
// 回傳值：
//   - *gameCenter: 初始化完成的遊戲中心服務結構指標。
func NewService(loginService login.Service, logger *slog.Logger, rdb *redis.Client) *gameCenter {
	s := &gameCenter{
		loginService: loginService,
		logger:       logger,
		redisClient:  rdb,
		games:        make(map[int]game.IGame),
		clientList:   make(map[string]game.GameClient),
	}

	// 如果有 Redis，啟動監聽器處理廣播指令
	if rdb != nil {
		go s.listenControlCommands()
	}

	return s
}

func (s *gameCenter) HandleConnect(client game.GameClient) {
	s.logger.Info("game service: client connected", "ip", client.GetIP())
	clientID := client.GetID()
	s.clientList[clientID] = client
}

func (s *gameCenter) HandleDisconnect(client game.GameClient) {
	s.logger.Info("game service: client disconnected", "ip", client.GetIP())
	clientID := client.GetID()
	delete(s.clientList, clientID)
	// 在這裡可以加入玩家離線的處理邏輯，例如從遊戲中移除
	player, _ := client.GetTag("player")
	if player != nil {
		s.leaveGame(*player.(*game.Player))
	}
}

func (s *gameCenter) HandleMessage(client game.GameClient, message []byte) {
	// 先解析 action 層
	var base struct {
		Action ActionType      `json:"action"`
		Data   json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(message, &base); err != nil {
		s.logger.Warn("failed to unmarshal message", "error", err, "ip", client.GetIP())
		client.Kick("invalid message format")
		return
	}

	s.logger.Info("message received", "action", base.Action, "ip", client.GetIP())

	switch base.Action {
	case Login:
		var payload loginPayload
		if err := json.Unmarshal(base.Data, &payload); err != nil {
			s.logger.Warn("invalid auth payload", "error", err, "ip", client.GetIP())
			return
		}
		s.handleLogin(client, payload.Sid)
		player, _ := client.GetTag("player")
		if player != nil {
			domainPlayer := *player.(*game.Player)
			s.joinGame(payload.GameID, domainPlayer)
		}
	case Play:
		var payload playPayload
		if err := json.Unmarshal(base.Data, &payload); err != nil {
			s.logger.Warn("invalid auth payload", "error", err, "ip", client.GetIP())
			return
		}
		s.handlePlay(client, payload.BetAmount)
	default:
		s.logger.Warn("unknown action", "action", base.Action, "ip", client.GetIP())
	}
}

func (s *gameCenter) handleLogin(gameClient game.GameClient, token string) {
	if token == "" {
		gameClient.Kick("auth failed: token is missing")
		return
	}

	player, err := s.loginService.Authenticate(token, gameClient)
	if err != nil {
		s.logger.Error("authentication failed", "error", err, "ip", gameClient.GetIP())
		gameClient.Kick("authentication failed")
		return
	}

	// 將驗證成功的 Player 物件附加到連線上
	gameClient.SetTag("player", player)
	s.logger.Info("client authenticated successfully", "playerID", player.ID, "playerName", player.Name, "ip", gameClient.GetIP())

	response := game.Envelope{
		Action:  "auth_success",
		Payload: fmt.Sprintf("{\"message\": \"authenticated successfully\", \"playerID\": \"%s\"}", player.ID),
	}
	player.SendMessage(response)
}

func (s *gameCenter) handlePlay(gameClient game.GameClient, betAmount decimal.Decimal) {
	player, _ := gameClient.GetTag("player")
	if player == nil {
		gameClient.Kick("Not Login")
		return
	}
	domainPlayer := player.(*game.Player)
	gameID, exists := domainPlayer.GetTag("game")
	if !exists {
		gameClient.Kick("Not in any game")
		return
	}
	realGameID := gameID.(int)
	game := s.games[realGameID]
	if game == nil {
		domainPlayer.Kick("game not found")
		return
	}
	game.Play(domainPlayer, betAmount)
}

func (s *gameCenter) RegisterGame(game game.IGame) {
	id := game.ID()
	s.games[id] = game
}

func (s *gameCenter) joinGame(gameID int, player game.Player) error {
	game := s.games[gameID]
	if game == nil {
		player.Kick("joinGame game not found")
		return fmt.Errorf("game not found")
	}
	game.AddPlayer(&player)
	player.SetTag("game", gameID)

	// Redis 全域計數
	if s.redisClient != nil {
		ctx := context.Background()
		key := fmt.Sprintf(RedisKeyPlayerCountPrefix, gameID)
		s.redisClient.Incr(ctx, key)
	}

	return nil
}

func (s *gameCenter) leaveGame(player game.Player) error {
	gameID, exists := player.GetTag("game")
	if !exists {
		return nil
	}
	realGameID := gameID.(int)
	game := s.games[realGameID]
	if game != nil {
		game.RemovePlayer(&player)
	}

	// Redis 全域計數
	if s.redisClient != nil {
		ctx := context.Background()
		key := fmt.Sprintf(RedisKeyPlayerCountPrefix, realGameID)
		s.redisClient.Decr(ctx, key)
	}

	return nil
}

// GetGames 取得所有遊戲的狀態列表。
//
// 此方法會嘗試從 Redis 取得所有遊戲的即時線上人數（跨服務實體加總）。
// 它使用了 Redis 的 SCAN 指令來避免阻塞，並透過 Pipeline 批次讀取計數。
//
// 參數說明：
//   - ctx: context.Context, 用於控制 Redis 請求的 Context。
//
// 回傳值：
//   - []GameInfo: 包含遊戲 ID 與當前線上人數的結構列表。
func (s *gameCenter) GetGames(ctx context.Context) []GameInfo {
	var (
		cursor uint64
		keys   []string
		games  []GameInfo
	)

	// 使用 SCAN 迭代搜尋所有符合 games:*:count 模式的 key
	for {
		var k []string
		var err error
		k, cursor, err = s.redisClient.Scan(ctx, cursor, "games:*:count", 100).Result()
		if err != nil {
			panic(err)
		}

		keys = append(keys, k...)

		if cursor == 0 {
			break
		}
	}

	pipe := s.redisClient.Pipeline()
	cmds := make([]*redis.StringCmd, 0, len(keys))
	for _, key := range keys {
		cmds = append(cmds, pipe.Get(ctx, key))
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		panic(err)
	}

	for i, cmd := range cmds {
		fmt.Println(keys[i], cmd.Val())
		parts := strings.Split(keys[i], ":")
		strGameID := parts[1]
		playerCount, _ := strconv.Atoi(cmd.Val())
		games = append(games, GameInfo{
			ID:          strGameID,
			PlayerCount: playerCount,
		})
	}
	return games
}

func (s *gameCenter) KickAll(ctx context.Context) error {
	if s.redisClient == nil {
		return fmt.Errorf("redis client is nil")
	}

	cmd := ControlCommand{
		Action: "kick_all",
	}
	payload, _ := json.Marshal(cmd)
	return s.redisClient.Publish(ctx, RedisChannelControl, payload).Err()
}

func (s *gameCenter) listenControlCommands() {
	ctx := context.Background()
	pubsub := s.redisClient.Subscribe(ctx, RedisChannelControl)
	defer pubsub.Close()

	ch := pubsub.Channel()
	s.logger.Info("listening to redis control commands")

	for msg := range ch {
		var cmd ControlCommand
		if err := json.Unmarshal([]byte(msg.Payload), &cmd); err != nil {
			s.logger.Warn("received invalid control command", "payload", msg.Payload)
			continue
		}

		s.logger.Info("control command received", "action", cmd.Action)

		switch cmd.Action {
		case "kick_all":
			s.handleGlobalKickAll()
		}
	}
}

func (s *gameCenter) handleGlobalKickAll() {
	s.logger.Warn("EXECUTING GLOBAL KICK ALL")
	for _, client := range s.clientList {
		client.Kick("api kick !")
	}
}
