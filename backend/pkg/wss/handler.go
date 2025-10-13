package wss

type Handler interface {
	OnConnect(client Client)
	OnDisconnect(client Client)
	OnMessage(client Client, message []byte)
}
