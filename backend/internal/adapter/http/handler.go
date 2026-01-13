package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joe_shih/slot-factory/internal/application/gamecenter"
	"github.com/joe_shih/slot-factory/internal/application/wallet"
)

// Handler 處理所有 REST API 請求。
type Handler struct {
	gameProvider  gamecenter.GameProvider
	adminProvider gamecenter.AdminProvider
	history       wallet.HistoryProvider
}

// NewHandler 建立一個新的 HTTP Handler 實例。
//
// 此函式採用介面隔離原則(ISP)注入依賴，確保 Handler 僅依賴它所需要的方法集。
//
// 參數說明：
//   - gp: gamecenter.GameProvider, 提供遊戲查詢功能。
//   - ap: gamecenter.AdminProvider, 提供管理員指令功能。
//   - hp: wallet.HistoryProvider, 提供錢包歷史查詢功能。
//
// 回傳值：
//   - *Handler: 初始化完成的 HTTP Handler 指標。
func NewHandler(gp gamecenter.GameProvider, ap gamecenter.AdminProvider, hp wallet.HistoryProvider) *Handler {
	return &Handler{
		gameProvider:  gp,
		adminProvider: ap,
		history:       hp,
	}
}

// HandleGetGames 回回傳所有註冊的遊戲列表。
func (h *Handler) HandleGetGames(c *gin.Context) {
	games := h.gameProvider.GetGames(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{
		"games": games,
	})
}

// HandleKickAll 處理全域踢除玩家的請求。
//
// 此 API 會廣播踢線指令到所有 WebSocket 伺服器實體。
// 方法：POST /api/v1/admin/kick_all
//
// 參數說明：
//   - c: *gin.Context, Gin 框架的 Context。
//
// 回傳值：
//   - JSON Response: 成功時回傳 200 OK，失敗時回傳 500 Internal Server Error。
func (h *Handler) HandleKickAll(c *gin.Context) {
	if err := h.adminProvider.KickAll(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "kick all signal sent"})
}

// HandleGetHistory 回傳玩家的交易歷史紀錄。
func (h *Handler) HandleGetHistory(c *gin.Context) {
	playerID := c.Query("playerID")
	if playerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "playerID is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	history, pErr := h.history.GetHistory(playerID, limit)
	if pErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": pErr.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"playerID": playerID,
		"history":  history,
	})
}
