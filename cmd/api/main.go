package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"go.uber.org/zap"

	"github.com/go-playground/validator/v10"
	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/myselfBZ/chatrix-v2/internal/auth"
	"github.com/myselfBZ/chatrix-v2/internal/db"
	"github.com/myselfBZ/chatrix-v2/internal/store"
	"github.com/olahol/melody"
)

const (
	userCtxValKey = "user"
)

func newApi(port int) *api {
	m := melody.New()
	a := &api{
		authConfig: authConfig{
			accessSecret:  "something",
			refreshSecret: "something",
			iss:           "chatrix",
			aud:           "chatrix",
		},
		port: port,
		mel:  m,
	}
	logger := zap.Must(zap.NewProduction(zap.AddCaller())).Sugar()
	defer logger.Sync()

	a.logger = logger
	a.validator = validator.New()

	a.auth = auth.NewJWTAuthenticator(
		a.authConfig.accessSecret,
		a.authConfig.refreshSecret,
		a.authConfig.aud,
		a.authConfig.iss,
	)

	dbUrl := os.Getenv("DB")

	if dbUrl == "" {
		panic("DB url is not set")
	}

	db, err := db.New(db.Config{
		Addr:        dbUrl,
		MaxConns:    15,
		MinConns:    15,
		MaxIdleTime: "15m",
	})

	if err != nil {
		panic(err)
	}

	a.storage = *store.NewStorage(db)

	m.HandleMessage(a.handleMessage)
	m.HandleConnect(a.handleConnect)
	m.HandleDisconnect(a.handleDisconnect)
	return a
}

type api struct {
	auth       auth.Authenticator
	authConfig authConfig
	port       int
	mel        *melody.Melody
	validator  *validator.Validate
	storage    store.Storage
	clients    sync.Map
	logger     *zap.SugaredLogger
}

// handlers

func (a *api) serve() error {
	e := echo.New()
	prodFrontEnd := os.Getenv("FRONTEND_URL")
	if prodFrontEnd == "" {
		panic("production front end is not set!")
	}
	localFrotnEndUrls := []string{
		"http://localhost:5173",
		"http://localhost:5174",
	}
	e.Use(middleware.RequestLogger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{localFrotnEndUrls[0], localFrotnEndUrls[1], prodFrontEnd},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
		},
		AllowCredentials: true,
	}))

	e.GET("/ws", a.handleWebSocket)
	e.POST("/auth/token", a.createTokenHandler)
	e.POST("/auth/users", a.createUserHandler)
	e.POST("/auth/refresh", a.refreshTokenHandler)
	
	e.GET("/protected", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"Pass": "OK",
		})
	}, a.AuthMiddleware)

	authenticatedRoutes := e.Group("/authenticated", a.AuthMiddleware)

	authenticatedRoutes.POST("/conversations", a.createConversationHandler)
	authenticatedRoutes.GET("/conversations/mine", a.getConversationsHandler)
	authenticatedRoutes.GET("/messages", a.getMessageHistoryHandler)
	authenticatedRoutes.POST("/users/search", a.searchUserHandler)

	return e.Start(fmt.Sprintf(":%d", a.port))
}

func main() {
	a := newApi(8080)
	slog.Info("Runnin'...")
	a.serve()
}
