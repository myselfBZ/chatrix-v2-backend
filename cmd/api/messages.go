package main

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	AKC_MSG_DELIVERED = "ACK_MSG_DELIVERED"
	WELCOME           = "WELCOME"
	ERR               = "ERR"
	CHAT              = "CHAT"
	ONLINE_PRESENCE   = "ONLINE_PRESENCE"
	OFFLINE_STATUS    = "OFFLINE_STATUS"
	MARK_READ         = "MARK_READ"
	MSG_READ          = "MSG_READ"
	CLIENT_CONN       = "CLIENT_CONN"
)

type IncomingEvent struct {
	MsgType string          `json:"type"`
	Message json.RawMessage `json:"message"`
}

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
	// User ids
	To   string `json:"to"`
	From string `json:"from"`

	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	TempID    string    `json:"temp_id"`
	ID        uuid.UUID `json:"id"`
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

type AcknowledgementMsgDelivered struct {
	RecieverID string    `json:"reciever_id"`
	TempID     string    `json:"temp_id"`
	CreatedAt  time.Time `json:"created_at"`
	ID         string    `json:"id"`
}

func (m *AcknowledgementMsgDelivered) message() {}

type MarkMsgRead struct {
	ConversationID string `json:"conversation_id"`
	MsgOwnerID     string `json:"msg_owner_id"`
}

func (m *MarkMsgRead) message() {}

type MsgRead struct {
	ConversationID string      `json:"conversation_id"`
	MessageIDs     []uuid.UUID `json:"message_ids"`
}

func (m *MsgRead) message() {}
