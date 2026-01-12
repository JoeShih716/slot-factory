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

// NewHandler 建立一個新的 HTTP Handler。..
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

// HandleKickAll 廣播踢出所有連線玩家。
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
