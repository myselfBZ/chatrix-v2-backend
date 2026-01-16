package main

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/myselfBZ/chatrix-v2/internal/queries"
	"github.com/myselfBZ/chatrix-v2/internal/store"
)

type createConversationPayload struct {
	// User1 is the intiator
	User1 string `json:"user1" validate:"required"`
	User2 string `json:"user2" validate:"required"`
}


type conversationResponse struct {
	UserData queries.GetConversationsByUserIDRow `json:"user_data"`
	UserIsOnline bool `json:"is_online"`
}

// WELL, WELL, we gotta fix this ASAP
func (m *conversationResponse) message() {}


func (a *api) getConversationsHandler(c echo.Context) error {
	user := c.Get(userCtxValKey).(queries.User)
	conversationsDB, err := a.storage.Conversations.GetByUserID(c.Request().Context(), user.ID)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			a.notFoundLog(c.Request().Method, c.Path(), err)
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		default:
			a.internalErrLog(c.Request().Method, c.Path(), err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
	}

 	convaersations := make([]conversationResponse, len(conversationsDB))

	for i, c := range conversationsDB {
		_, isOnline := a.clients.Load(c.ID.String())
		convaersations[i] = conversationResponse{
			UserData: c,
			UserIsOnline: isOnline,
		}
	}

	return c.JSON(http.StatusOK, convaersations)
}

func (a *api) createConversationHandler(c echo.Context) error {
	user := c.Get(userCtxValKey).(queries.User)
	var payload createConversationPayload
	if err := c.Bind(&payload); err != nil {
		a.badRequestLog(c.Request().RequestURI, c.Path(), err)
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	if err := a.validator.Struct(payload); err != nil {
		a.badRequestLog(c.Request().RequestURI, c.Path(), err)
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	user1ValidID, _ := uuid.Parse(payload.User1)
	user2ValidID, _ := uuid.Parse(payload.User2)

	conversation, err := a.storage.Conversations.Create(c.Request().Context(), queries.CreateConversationParams{
		User1: user1ValidID,
		User2: user2ValidID,
	})

	if err != nil {
		switch err {
		case store.ErrAlreadyExists:
			a.conflictLog(c.Request().RequestURI, c.Path(), err)
			return echo.NewHTTPError(http.StatusConflict)
		case store.ErrConstraintMessage:
			a.badRequestLog(c.Request().RequestURI, c.Path(), err)
			return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
		default:
			a.internalErrLog(c.Request().RequestURI, c.Path(), err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
	}

	_, isOnline := a.clients.Load(user1ValidID.String())

	go a.notifyConversationCreation(user2ValidID, conversationResponse{
		UserData: queries.GetConversationsByUserIDRow{
			ID: user1ValidID,
			LastSeen: user.LastSeen,
			ConversationID: conversation.ID,
			Username: user.Username,
		},
		UserIsOnline: isOnline,
	})

	return c.JSON(http.StatusOK, conversation)
}
