package main

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/myselfBZ/chatrix-v2/internal/store"
	"golang.org/x/crypto/bcrypt"

	"github.com/myselfBZ/chatrix-v2/internal/queries"
)

type authConfig struct {
	accessSecret, refreshSecret string
	aud                        string
	iss                         string
}

type loginPayload struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type userPayload struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	Email    string `json:"email" validate:"required"`
}

type tokenEnvelope struct {
	User  *queries.User `json:"user"`
	Token string      `json:"access_token"`
}

func (a *api) refreshTokenHandler(c echo.Context) error {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		a.unauthorizedLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusUnauthorized, "no refresh token")
	}

	token, err := a.auth.ValidateRefreshToken(cookie.Value)
	if err != nil {
		c.SetCookie(&http.Cookie{
			Name:   "refresh_token",
			Value:  "",
			MaxAge: -1,
		})
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid refresh token")
	}

	userID, err := a.auth.ExtractUserID(token)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token claims")
	}

	validUUID, err := uuid.Parse(userID)
	
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid uuid")
	}

	user, err := a.storage.Users.GetByID(c.Request().Context(), validUUID)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "user not found")
	}

	accessToken, err := a.auth.GenerateAccessToken(user.ID.String())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate tokens")
	}

	return c.JSON(http.StatusOK, &tokenEnvelope{
		Token: accessToken,
		User:  &user,
	})
}

func (a *api) createTokenHandler(c echo.Context) error {
	var payload loginPayload

	if err := c.Bind(&payload); err != nil {
		return err
	}

	if err := a.validator.Struct(payload); err != nil {
		a.badRequestLog(c.Request().RequestURI, c.Path(), err)
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	user, err := a.storage.Users.GetByEmail(c.Request().Context(), payload.Email)

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

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(payload.Password)); err != nil {
		a.internalErrLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

	tokens, err := a.auth.GenerateTokenPair(user.ID.String(), make(map[string]any))

	if err != nil {
		a.internalErrLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   7 * 24 * 60 * 60, // 7 days
	})

	return c.JSON(http.StatusOK, &tokenEnvelope{
		Token: tokens.AccessToken,
		User:  &user,
	})
}

func (a *api) createUserHandler(c echo.Context) error {
	var payload userPayload
	if err := c.Bind(&payload); err != nil {
		return err
	}

	if err := a.validator.Struct(payload); err != nil {
		a.badRequestLog(c.Request().RequestURI, c.Path(), err)
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	user := queries.CreateUserParams{
		Email:    payload.Email,
		Username: payload.Username,
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)

	if err != nil {
		a.internalErrLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

	user.PasswordHash = string(hash)

	dbUser, err := a.storage.Users.Create(c.Request().Context(), user)

	if err != nil {
		switch err {
		case store.ErrAlreadyExists:
			a.conflictLog(c.Request().Method, c.Path(), err)
			return echo.NewHTTPError(http.StatusConflict, err)
		default:
			a.internalErrLog(c.Request().Method, c.Path(), err)
			return echo.NewHTTPError(http.StatusInternalServerError)

		}

	}

	tokens, err := a.auth.GenerateTokenPair(dbUser.ID.String(), make(map[string]any))

	if err != nil {
		a.internalErrLog(c.Request().Method, c.Path(), err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, tokenEnvelope{
		User: &dbUser,
		Token: tokens.AccessToken,
	})
}

