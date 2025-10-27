package gamecenter

import "github.com/shopspring/decimal"

// loginPayload Login專用資料結構
type loginPayload struct {
	Sid    string `json:"sid"`
	GameID int    `json:"gameId"`
}

// playPayload Play專用結構
type playPayload struct {
	BetAmount decimal.Decimal `json:"betAmount"`
}
