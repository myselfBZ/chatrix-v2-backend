package main

import (
	"net/http"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/myselfBZ/chatrix-v2/internal/queries"
	"github.com/myselfBZ/chatrix-v2/internal/store"
)

func (a *api) getMessageHistoryHandler(c echo.Context) error {
	id := c.QueryParam("with_id")
	validUUID, err := uuid.Parse(id)
	if err != nil {
		a.badRequestLog(c.Request().RequestURI, c.Path(), err)
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	user := c.Get(userCtxValKey).(queries.User)

	conversation, err := a.storage.Conversations.GetByMembers(c.Request().Context(), queries.GetConversationByMembersParams{
		User1: user.ID,
		User2: validUUID,
	})

	if err != nil {
		switch err {
		case store.ErrNotFound:
			a.notFoundLog(c.Request().Method, c.Path(), err)
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		default:
			a.internalErrLog(c.Request().Method, c.Path(), err)
			return echo.NewHTTPError(http.StatusUnauthorized)
		}
	}

	msgs, err := a.storage.Messages.GetByConversationID(c.Request().Context(), queries.GetMessagesByConversationIDParams{
		ConversationID: conversation.ID,
		Offset: 0,
		Limit: 100,
	})

	return c.JSON(http.StatusOK, msgs)
}
