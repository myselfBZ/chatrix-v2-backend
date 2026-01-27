package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/myselfBZ/chatrix-v2/internal/queries"
	"github.com/olahol/melody"
)

const (
	userIDSessionKey = "user_id"
	authSessionKey   = "authenticated"
)

type authPayload struct {
	Message struct {
		Token string `json:"token"`
	} `json:"message"`
}

func (a *api) handleDisconnect(s *melody.Session) {
	userID, ok := s.Get(userIDSessionKey)
	if !ok {
		return
	}
	userIDString := userID.(string)
	a.clients.Delete(userIDString)
	validUUID, _ := uuid.Parse(userIDString)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	a.storage.Users.UpdateLastSeen(ctx, validUUID)
	a.broadCastOfflineStatus(validUUID)
}

func (a *api) handleConnect(s *melody.Session) {
	go func() {
		time.Sleep(5 * time.Second)
		if !a.isSessionAuthenticated(s) {
			s.WebsocketConnection().Close()
		}
	}()
}

func (a *api) handleMessage(s *melody.Session, msg []byte) {

	if !a.isSessionAuthenticated(s) {
		// Session not authenticated
		// should be authenticated
		user, err := a.authenticateSession(msg)

		if err != nil {
			// todo better error handling
			jsonErr, _ := json.Marshal(Wrapper{
				MsgType: ERR,
				Message: &Err{
					Reason: "we couldn't authenticate you",
					Code:   http.StatusUnauthorized,
				},
			})
			s.CloseWithMsg(jsonErr)
			return
		}
		s.Set(userIDSessionKey, user.ID.String())
		s.Set(authSessionKey, true)

		a.clients.Store(user.ID.String(), s)
		welcome, _ := json.Marshal(Wrapper{
			MsgType: WELCOME,
			Message: &Welcome{},
		})
		s.Write(welcome)
		a.broadcaseOnlineStatus(user.ID)
		return
	}

	var event IncomingEvent
	if err := json.Unmarshal(msg, &event); err != nil {
		a.logger.Error("couldn't Unmarshal Incoming event", err)
		writeJSONErr(s, &Err{
			Reason: "invalid payload",
			Code: http.StatusUnprocessableEntity,
		})
		return
	}

	a.mapIncomingEventToHandler(s, &event)
}


func (a *api) mapIncomingEventToHandler(s *melody.Session, event *IncomingEvent) {
	switch event.MsgType {
	case CHAT:
		var payload ChatMsg
		if err := json.Unmarshal(event.Message, &payload); err != nil {
			a.logger.Error("couldn't Unmarshal ChatMsg event", err)
			writeJSONErr(s, &Err{
				Reason: "invalid payload",
				Code: http.StatusUnprocessableEntity,
			})
			return
		}

		a.handleChatMessage(s, &payload)
	case MARK_READ:
		var payload MarkMsgRead

		if err := json.Unmarshal(event.Message, &payload); err != nil {
			writeJSONErr(s, &Err{
				Reason: "invalid payload",
				Code: http.StatusUnprocessableEntity,
			})
		}


		a.handleMarkMsgRead(s, &payload)
	case TYPING:
		var payload Typing
		if err := json.Unmarshal(event.Message, &payload); err != nil {
			writeJSONErr(s, &Err{
				Reason: "invalid payload",
				Code: http.StatusUnprocessableEntity,
			})
		}

		a.handleTyping(s, &payload)
	case STOPPED_TYPING:
		var payload StoppedTyping
		if err := json.Unmarshal(event.Message, &payload); err != nil {
			writeJSONErr(s, &Err{
				Reason: "invalid payload",
				Code: http.StatusUnprocessableEntity,
			})
		}
		a.handleStoppedTyping(s, &payload)
	}
}


func (a *api) handleMarkMsgRead(s *melody.Session, msg *MarkMsgRead) {
	conversationID, err := uuid.Parse(msg.ConversationID)
	if err != nil {
		return
	}

	ownerID, err := uuid.Parse(msg.MsgOwnerID)
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	msgIds ,err := a.storage.Messages.MarkAsRead(ctx, queries.MarkMessagesAsReadParams{
		ConversationID: conversationID,
		SenderID: ownerID,
	}) 

	if err != nil {
		return
	}

	session, ok := a.getSession(ownerID)
	if !ok {
		return
	}


	writeJSONMsg(session, Wrapper{
		MsgType: MSG_READ,
		Message: &MsgRead{
			ConversationID: msg.ConversationID,
			MessageIDs: msgIds,
		},
	})
}



func (a *api) handleChatMessage(s *melody.Session, msg *ChatMsg) {
	fromUUID, err := uuid.Parse(msg.From)

	if err != nil {
		writeJSONErr(s, &MessageErr{
			TempID: msg.TempID,
			Reason: "invalid user UUID",
		})
		return
	}

	toUUID, err := uuid.Parse(msg.To)

	if err != nil {
		writeJSONErr(s, &MessageErr{
			Reason: "invalid UUID",
			TempID: msg.TempID,
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbMsg, err := a.storage.Messages.Create(ctx, queries.CreateMessageParams{
		SenderID: fromUUID,
		User2:    toUUID,
		Content:  msg.Content,
	})

	if err != nil {
		writeJSONErr(s, 
			&MessageErr{
				Reason: "message couldn't be created",
				TempID: msg.TempID,
			},
		)
		return
	} else {
		go writeJSONMsg(s, Wrapper{
			MsgType: AKC_MSG_DELIVERED,
			Message: &AcknowledgementMsgDelivered{
				RecieverID: msg.To,
				CreatedAt: dbMsg.CreatedAt.Time,
				TempID:    msg.TempID,
				ID:        dbMsg.ID.String(),
			},
		})
	}


	msg.CreatedAt = dbMsg.CreatedAt.Time
	msg.ID = dbMsg.ID

	session, ok := a.getSession(toUUID)

	if !ok {
		return
	}

	writeJSONMsg(session, Wrapper{
		MsgType: CHAT,
		Message: msg,
	})
}

func (a *api) handleStoppedTyping(s *melody.Session, msg *StoppedTyping) {
	toUUID, err := uuid.Parse(msg.To)
	if err != nil {
		writeJSONErr(s, &Err{
			Reason: "invalid UUID",
			Code: http.StatusUnprocessableEntity,
		})
		return
	}

	toSession, ok := a.getSession(toUUID)

	if !ok {
		return
	}

	writeJSONMsg(toSession, Wrapper{
		MsgType: STOPPED_TYPING,
		Message: msg,
	})
}


func (a *api) handleTyping(s *melody.Session, msg *Typing) {
	toUUID, err := uuid.Parse(msg.To)
	if err != nil {
		writeJSONErr(s, &Err{
			Reason: "invalid UUID",
			Code: http.StatusUnprocessableEntity,
		})
		return
	}

	toSession, ok := a.getSession(toUUID)

	if !ok {
		return
	}

	writeJSONMsg(toSession, Wrapper{
		MsgType: TYPING,
		Message: msg,
	})
}

func (a *api) handleWebSocket(c echo.Context) error {
	a.mel.HandleRequest(c.Response().Writer, c.Request())
	return nil
}

func (a *api) authenticateSession(msg []byte) (queries.User, error) {
	var payload authPayload
	if err := json.Unmarshal(msg, &payload); err != nil {
		return queries.User{}, err
	}
	token := payload.Message.Token
	jwtToken, err := a.auth.ValidateAccessToken(token)
	if err != nil {
		return queries.User{}, err
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)

	if !ok {
		return queries.User{}, errors.New("not jwt.MapClaims")
	}

	userID, ok := claims["sub"].(string)

	if !ok {
		return queries.User{}, errors.New("invalid user id")
	}

	validUUID, err := uuid.Parse(userID)

	if err != nil {
		return queries.User{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	user, err := a.storage.Users.GetByID(ctx, validUUID)

	if err != nil {
		return queries.User{}, err
	}

	return user, nil
}

func (a *api) isSessionAuthenticated(s *melody.Session) bool {
	isAuth, ok := s.Get(authSessionKey)
	if !ok || isAuth == nil {
		return false
	}

	authenticated, ok := isAuth.(bool)
	return ok && authenticated
}

func writeJSONMsg(s *melody.Session, payload Wrapper) error {
	jsonData, _ := json.Marshal(payload)

	return s.Write(jsonData)
}

func writeJSONErr(s *melody.Session, err Message) error {
	return writeJSONMsg(s, Wrapper{
		MsgType: ERR,
		Message: err,
	})
}


func (a *api) getSession(id uuid.UUID) (*melody.Session, bool) {
	sessionAny, ok := a.clients.Load(id.String())
	if !ok {
		return nil, ok
	}
	session := sessionAny.(*melody.Session)
	return session, ok
}
