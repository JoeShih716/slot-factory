package gamecenter

type ActionType string

const (
	// --- Client to Server Actions ---
	Login ActionType = "login"
	Play  ActionType = "play"
)
