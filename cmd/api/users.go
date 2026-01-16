package main

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/myselfBZ/chatrix-v2/internal/queries"
	"github.com/myselfBZ/chatrix-v2/internal/store"
)

type searchUserResponse struct {
	IsOnline bool      `json:"is_online"`
	Username string    `json:"username"`
	LastSeen time.Time `json:"last_seen"`
	ID       uuid.UUID `json:"id"`
}

type searchPayload struct {
	Query string `json:"query"`
}

func (a *api) searchUserHandler(c echo.Context) error {
	var payload searchPayload
	user := c.Get(userCtxValKey).(queries.User)
	if err := c.Bind(&payload); err != nil {
		a.badRequestLog(c.Request().RequestURI, c.Path(), err)
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	users, err := a.storage.Users.Search(c.Request().Context(), user.Username, payload.Query)
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

	convaersations := make([]searchUserResponse, len(users))

	for i, c := range users {
		_, isOnline := a.clients.Load(c.ID.String())
		convaersations[i] = searchUserResponse{
			IsOnline: isOnline,
			LastSeen: c.LastSeen.Time,
			ID: c.ID,
			Username: c.Username,
		}
	}

	return c.JSON(http.StatusOK, convaersations)
}
