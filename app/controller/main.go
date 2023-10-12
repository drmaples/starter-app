package controller

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	slogecho "github.com/samber/slog-echo"
)

// FIXME: move this
const schema = "public"

// Initialize sets up the controller layer
func Initialize() *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Validator = newValidator()
	e.Use(slogecho.New(slog.Default()))
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 30 * time.Second,
	}))

	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]any{"howdy": "there"})
	})

	e.GET("/user", handleListUsers)
	e.GET("/user/:id", handleGetUser)
	e.POST("/user", handleCreateUser)

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
