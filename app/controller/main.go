package controller

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	slogecho "github.com/samber/slog-echo"

	"github.com/drmaples/starter-app/app/platform"
)

// GetServerAddress is address where server runs
func GetServerAddress() string {
	return fmt.Sprintf("%s:%d", platform.Config().ServerURL, platform.Config().ServerPort)
}

// GetServerBindAddress is bind address for echo server
func GetServerBindAddress() string {
	return fmt.Sprintf(":%d", platform.Config().ServerPort)
}

// Initialize sets up the controller layer
func Initialize() *echo.Echo {
	e := echo.New()
	if platform.Config().Environment != "dev" {
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
				SigningKey: []byte(platform.Config().JWTSignKey),
				NewClaimsFunc: func(c echo.Context) jwt.Claims {
					return new(jwtCustomClaims)
				},
			}),
		)
		restricted.GET("/user", handleListUsers)
		restricted.GET("/user/:id", handleGetUser)
		restricted.POST("/user", handleCreateUser)
	}

	// list routes in use like gin. keep at bottom
	if platform.Config().Environment == "dev" {
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
