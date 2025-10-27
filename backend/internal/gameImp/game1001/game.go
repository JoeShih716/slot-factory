package game1001

import (
	"log/slog"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/joe_shih/slot-factory/internal/domain/game"
	"github.com/shopspring/decimal"
)

// betInfo 記錄了單一玩家的下注資訊。
type betInfo struct {
	betAmount decimal.Decimal
}

// gamePlayer 擴展了 domain.Player，加入了遊戲內的特定資料，如下注資訊。
type gamePlayer struct {
	*game.Player
	betInfo betInfo
}

// Game 實作一個多人輪盤遊戲 (game.IGame 介面)。
// 遊戲邏輯：每隔一段時間開獎，開出數字 1~10，開中 1 且有下注的玩家贏得10倍彩金。
type Game struct {
	id        int
	players   map[string]*gamePlayer // 使用玩家 ID 作為 key
	mu        sync.RWMutex           // 使用讀寫鎖保護 players map
	state     state
	countdown int
	ticker    *time.Ticker  // 遊戲主循環的計時器
	stopCh    chan struct{} // 用於停止遊戲主循環
	logger    *slog.Logger
}

// NewGame 創建一個新的 1001 輪盤遊戲實例。
func NewGame(logger *slog.Logger) game.IGame {
	game := &Game{
		id:      1001,
		players: make(map[string]*gamePlayer),
		state:   StateWaiting,
		ticker:  time.NewTicker(1 * time.Second),
		stopCh:  make(chan struct{}),
		logger:  logger.With("gameID", 1001),
	}
	game.startLoop()
	return game
}

// ID 返回遊戲的唯一標識符。
func (g *Game) ID() int {
	return g.id
}

// AddPlayer 將一個新玩家加入遊戲，並觸發狀態同步。
func (g *Game) AddPlayer(player *game.Player) {
	g.mu.Lock()
	// 1. 建立遊戲玩家實例並加入列表
	gp := &gamePlayer{
		Player:  player,
		betInfo: betInfo{betAmount: decimal.Zero},
	}
	g.players[player.ID] = gp
	g.logger.Info("player added", "playerID", player.ID)

	// 2. 準備新玩家和現有玩家的資訊
	newPlayerInfo := PlayerInfo{ID: gp.ID, Name: gp.Name, BetAmount: gp.betInfo.betAmount}

	// 複製一份當前玩家列表以在解鎖後使用
	playersToNotify := make([]*gamePlayer, 0, len(g.players)-1)
	currentPlayerListForNewPlayer := make([]PlayerInfo, 0, len(g.players)-1)
	for _, p := range g.players {
		if p.ID != player.ID { // 不包含新玩家自己
			playersToNotify = append(playersToNotify, p)
			currentPlayerListForNewPlayer = append(currentPlayerListForNewPlayer, PlayerInfo{ID: p.ID, Name: p.Name, BetAmount: p.betInfo.betAmount})
		}
	}
	currentState := g.state
	currentCountdown := g.countdown
	g.mu.Unlock() // !!! 在廣播前解鎖

	// 3. 廣播 player_joined 給所有已在房內的玩家
	g.broadcast(playersToNotify, game.Envelope{
		Action:  string(ActionPlayerJoined),
		Payload: PayloadPlayerJoined{Player: newPlayerInfo},
	})

	// 4. 發送 player_list 給新加入的玩家
	player.SendMessage(game.Envelope{
		Action:  string(ActionPlayerList),
		Payload: PayloadPlayerList{Players: currentPlayerListForNewPlayer},
	})

	// 5. 發送當前遊戲狀態給新玩家
	player.SendMessage(game.Envelope{
		Action: string(ActionStateUpdate),
		Payload: PayloadStateUpdate{
			State:     currentState,
			Countdown: currentCountdown,
		},
	})
}

// RemovePlayer 從遊戲中移除一個玩家。
func (g *Game) RemovePlayer(player *game.Player) {
	g.mu.Lock()
	delete(g.players, player.ID)
	// 複製玩家列表用於廣播
	allPlayers := g.getAllPlayers_unsafe()
	g.mu.Unlock() // 解鎖

	// 廣播玩家離開的訊息
	g.broadcast(allPlayers, game.Envelope{
		Action:  string(ActionPlayerLeft),
		Payload: PayloadPlayerLeft{PlayerID: player.ID},
	})
	g.logger.Info("player removed", "playerID", player.ID)
}

// Play 處理玩家的下注請求。
func (g *Game) Play(player *game.Player, betAmount decimal.Decimal) {
	g.mu.Lock()
	gamePlayer, ok := g.players[player.ID]
	if !ok {
		g.mu.Unlock()
		return
	}

	if g.state != StateBetting {
		g.mu.Unlock() // 解鎖後再發訊息
		gamePlayer.SendMessage(game.Envelope{Action: string(ActionBetResult), Payload: PayloadBetResult{Success: false, Error: "not in betting state"}})
		return
	}
	if betAmount.LessThanOrEqual(decimal.Zero) {
		g.mu.Unlock() // 解鎖後再發訊息
		gamePlayer.SendMessage(game.Envelope{Action: string(ActionBetResult), Payload: PayloadBetResult{Success: false, Error: "bet amount must be positive"}})
		return
	}

	// 更新玩家下注總額
	gamePlayer.betInfo.betAmount = gamePlayer.betInfo.betAmount.Add(betAmount)

	// 準備廣播資訊
	betResultPayload := PayloadBetResult{
		Success:  true,
		TotalBet: gamePlayer.betInfo.betAmount,
	}
	playerBetPayload := PayloadPlayerBet{
		PlayerID:  gamePlayer.ID,
		BetAmount: betAmount,
		TotalBet:  gamePlayer.betInfo.betAmount,
	}

	// 複製玩家列表用於廣播
	allPlayers := g.getAllPlayers_unsafe()
	g.mu.Unlock() // !!! 解鎖

	// 回傳個人下注結果
	gamePlayer.SendMessage(game.Envelope{
		Action:  string(ActionBetResult),
		Payload: betResultPayload,
	})

	// 廣播玩家下注活動
	g.broadcast(allPlayers, game.Envelope{
		Action:  string(ActionPlayerBet),
		Payload: playerBetPayload,
	})
}

// startLoop 啟動遊戲的主循環。
func (g *Game) startLoop() {
	go func() {
		g.logger.Info("game loop started")
		g.tick() // 立即觸發一次

		for {
			select {
			case <-g.ticker.C:
				g.tick()
			case <-g.stopCh:
				g.logger.Info("game loop stopped")
				return
			}
		}
	}()
}

// tick 是遊戲的核心驅動，每秒被調用一次。
func (g *Game) tick() {
	g.mu.Lock()
	g.countdown--
	if g.countdown > 0 {
		g.mu.Unlock()
		return
	}

	// 倒數結束，切換狀態
	switch g.state {
	case StateWaiting:
		g.setState_unsafe(StateBetting, 10)
	case StateBetting:
		// rollWheel 包含自己的鎖管理，所以這裡要先解鎖
		g.mu.Unlock()
		g.rollWheel()
		// 重新獲取鎖以進入下一個狀態
		g.mu.Lock()
		g.setState_unsafe(StateWaiting, 3)
	}
	g.mu.Unlock()
}

// setState_unsafe 切換遊戲狀態，此為內部方法，必須在持有寫鎖的情況下調用。
func (g *Game) setState_unsafe(next state, duration int) {
	g.state = next
	g.countdown = duration
	g.logger.Info("state changed", "state", next, "countdown", duration)

	// 準備廣播訊息
	msg := game.Envelope{
		Action: string(ActionStateUpdate),
		Payload: PayloadStateUpdate{
			State:     g.state,
			Countdown: g.countdown,
		},
	}
	allPlayers := g.getAllPlayers_unsafe()

	// 在 goroutine 中廣播，避免阻塞當前狀態機
	go g.broadcast(allPlayers, msg)
}

// rollWheel 執行開獎邏輯並廣播結果。
func (g *Game) rollWheel() {
	g.logger.Info("rolling wheel")
	number := rand.IntN(10) + 1
	isWin := number == 1

	// 準備廣播訊息
	openingMsg := game.Envelope{
		Action:  string(ActionOpening),
		Payload: PayloadOpening{Number: number},
	}

	g.mu.Lock() // 加鎖以安全地遍歷和修改玩家
	// 遍歷玩家，計算並發送輸贏結果
	for _, p := range g.players {
		if p.betInfo.betAmount.GreaterThan(decimal.Zero) {
			betAmount := p.betInfo.betAmount
			winAmount := decimal.Zero
			if isWin {
				winAmount = betAmount.Mul(decimal.NewFromInt(10))
			}

			// 在 goroutine 中發送個人訊息，避免阻塞
			go p.SendMessage(game.Envelope{
				Action: string(ActionWinResult),
				Payload: PayloadWinResult{
					BetAmount: betAmount,
					WinAmount: winAmount,
				},
			})
			// 重置玩家下注額
			p.betInfo.betAmount = decimal.Zero
		}
	}
	allPlayers := g.getAllPlayers_unsafe()
	g.mu.Unlock() // 解鎖

	// 廣播開獎號碼
	g.broadcast(allPlayers, openingMsg)
}

// broadcast 將訊息發送給指定的玩家列表。
func (g *Game) broadcast(players []*gamePlayer, message game.Envelope) {
	for _, p := range players {
		go p.SendMessage(message) // 使用 goroutine 非阻塞地發送
	}
}

// getAllPlayers_unsafe 返回所有玩家的列表，呼叫前必須持有鎖。
func (g *Game) getAllPlayers_unsafe() []*gamePlayer {
	players := make([]*gamePlayer, 0, len(g.players))
	for _, p := range g.players {
		players = append(players, p)
	}
	return players
}

// Stop 停止遊戲的主循環。
func (g *Game) Stop() {
	close(g.stopCh)
	g.ticker.Stop()
}

// 確保 Game 類型在編譯時期就實現了 IGame 接口。
var _ game.IGame = (*Game)(nil)
