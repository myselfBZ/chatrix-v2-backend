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
	a.storage.Users.UpdateLastSeen(context.TODO(), validUUID)
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

	// TODO: this not always a chat message
	payload := Wrapper{
		Message: &ChatMsg{},
	}

	if err := json.Unmarshal(msg, &payload); err != nil {
		// TODO: do something with the error
		return
	}

	chatMsg, ok := payload.Message.(*ChatMsg)

	if !ok {
		// TODO:  violation, do something with it
		return
	}

	fromUUID, err := uuid.Parse(chatMsg.From)

	if err != nil {
		// TODO: do something with the error
		return
	}

	toUUID, err := uuid.Parse(chatMsg.To)

	if err != nil {
		// TODO: do something with the error
		return
	}

	dbMsg, err := a.storage.Messages.Create(context.TODO(), queries.CreateMessageParams{
		SenderID: fromUUID,
		User2:    toUUID,
		Content:  chatMsg.Content,
	})


	if err != nil {
		// TODO
		writeJSONMsg(s, Wrapper{
			MsgType: ERR,
			Message: &Err{
				Reason: "message couldn't be created",
				// this is not suitable..... but uhm it is okay
				Code: http.StatusBadRequest,
			},
		})
	}

	go writeJSONMsg(s, Wrapper{
		MsgType: AKC_MSG_DELIVERED,
		Message: &AcknowledgementMsgDelivered{
			RecieverID: chatMsg.To,
			CreatedAt: dbMsg.CreatedAt.Time,
			TempID:    chatMsg.TempID,
			ID:        dbMsg.ID.String(),
		},
	})

	chatMsg.CreatedAt = dbMsg.CreatedAt.Time
	chatMsg.ID = dbMsg.ID

	targetUser, ok := a.clients.Load(chatMsg.To)
	if !ok {
		return
	}

	session := targetUser.(*melody.Session)

	writeJSONMsg(session, Wrapper{
		MsgType: CHAT,
		Message: chatMsg,
	})
}

func (a *api) handleWebSocket(c echo.Context) error {
	a.mel.HandleRequest(c.Response().Writer, c.Request())
	return nil
}

// TODO: anything in this file that's named authenticate* should be replaced
// to authorize

// Edit: 2026-01-11
// well, maybe not
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

	// TODO fix this context
	user, err := a.storage.Users.GetByID(context.TODO(), validUUID)

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
