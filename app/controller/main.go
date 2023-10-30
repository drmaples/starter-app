package controller

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/caarlos0/env/v9"
	"github.com/go-playground/validator/v10"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	slogecho "github.com/samber/slog-echo"
)

var envCfg config

// FIXME: find a better location for this
type config struct {
	GoogleClientID     string `env:"GOOGLE_CLIENT_ID,required"`
	GoogleClientSecret string `env:"GOOGLE_CLIENT_SECRET,required"`
	Environment        string `env:"ENVIRONMENT" envDefault:"dev"`
	ServerURL          string `env:"SERVER_URL" envDefault:"http://localhost"`
	ServerPort         int    `env:"SERVER_PORT" envDefault:"8000"`
	JWTSignKey         string `env:"JWT_SIGN_KEY" envDefault:"my-secret"` // do not want default
}

// GetServerAddress is address where server runs
func GetServerAddress() string {
	return fmt.Sprintf("%s:%d", envCfg.ServerURL, envCfg.ServerPort)
}

// GetServerBindAddress is bind address for echo server
func GetServerBindAddress() string {
	return fmt.Sprintf(":%d", envCfg.ServerPort)
}

// Initialize sets up the controller layer
func Initialize() *echo.Echo {
	if err := env.Parse(&envCfg); err != nil {
		panic(err)
	}

	e := echo.New()
	if envCfg.Environment != "dev" {
		e.HideBanner = true
		e.HidePort = true
	}
	e.Validator = newValidator()
	e.Use(slogecho.New(slog.Default()))
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 30 * time.Second,
	}))

	unrestricted := e.Group("")
	{
		unrestricted.GET("/", func(c echo.Context) error {
			return c.JSON(http.StatusOK, map[string]any{"howdy": "there"})
		})
		unrestricted.GET("/favicon.ico", func(c echo.Context) error { return nil }) // avoids 404 errors in the browser
		unrestricted.GET("/login", handleLogin)
		unrestricted.GET(oauthCallbackURL, handleOauthCallback)
	}

	restricted := e.Group("")
	{
		restricted.Use(
			echojwt.WithConfig(echojwt.Config{
				SigningKey: []byte(envCfg.JWTSignKey),
			}),
		)
		restricted.GET("/user", handleListUsers)
		restricted.GET("/user/:id", handleGetUser)
		restricted.POST("/user", handleCreateUser)
	}

	// list routes in use like gin. keep at bottom
	if envCfg.Environment == "dev" {
		for _, r := range e.Routes() {
			fmt.Printf("[%-5s] %-35s --> %s\n", strings.ToUpper(r.Method), r.Path, r.Name)
		}
	}
	return e
}

type customValidator struct {
	validator *validator.Validate
}

func newValidator() echo.Validator {
	return &customValidator{validator: validator.New()}
}

func (cv *customValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}
