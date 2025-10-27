package game

import "github.com/shopspring/decimal"

type IGame interface {
	ID() int
	AddPlayer(player *Player)
	RemovePlayer(player *Player)
	Play(player *Player, betAmount decimal.Decimal)
}