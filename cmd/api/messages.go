package main

import "time"

const (
	WELCOME         = "WELCOME"
	ERR             = "ERR"
	SET_NAME        = "SET_NAME"
	CHAT            = "CHAT"
	ONLINE_PRESENCE = "ONLINE_PRESENCE"
	OFFLINE_STATUS  = "OFFLINE_STATUS"
	CLIENT_CONN     = "CLIENT_CONN"
)

type Message interface {
	message()
}

type Wrapper struct {
	MsgType string  `json:"type"`
	Message Message `json:"message"`
}

type InitialServerMsg struct {
	Name    string   `json:"name"`
	Clients []string `json:"clients"`
}

func (m *InitialServerMsg) message() {}

type ChatMsg struct {
	To   string `json:"to"`
	From string `json:"from"`

	Content string `json:"content"`
}

func (m *ChatMsg) message() {}

type ClientDisconnected struct {
	Name string `json:"name"`
}

func (m *ClientDisconnected) message() {}

type ClientConnected struct {
	Name string `json:"name"`
}

func (m *ClientConnected) message() {}

type Err struct {
	Reason string `json:"reason"`
	Code   int    `json:"code"`
}

func (m *Err) message() {}

type Welcome struct{}

func (m *Welcome) message() {}

// a message to notify other users that a certain user has gone online
type OnlinePresence struct {
	UserID string `json:"user_id"`
}

func (m *OnlinePresence) message() {}

type OfflineStatus struct {
	UserID   string    `json:"user_id"`
	LastSeen time.Time `json:"last_seen"`
}

func (m *OfflineStatus) message() {}
