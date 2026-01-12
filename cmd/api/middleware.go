package main

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/google/uuid"
	"net/http"
	"strings"
)

func (app *api) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "authorization header is missing")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return echo.NewHTTPError(http.StatusUnauthorized, "authorization header is malformed")
		}

		token := parts[1]
		jwtToken, err := app.auth.ValidateAccessToken(token)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
		}

		claims, ok := jwtToken.Claims.(jwt.MapClaims)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid claims")
		}

		userID, ok := claims["sub"].(string)

		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid subject claim")
		}

		validUUID, err := uuid.Parse(userID)

		if err != nil {
			return err
		}

		user, err := app.storage.Users.GetByID(c.Request().Context(), validUUID)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "user not found")
		}

		c.Set(userCtxValKey, user)

		return next(c)
	}
}
