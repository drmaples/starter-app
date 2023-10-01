package controller

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	slogecho "github.com/samber/slog-echo"
)

// Initialize sets up the controller layer
func Initialize() *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(slogecho.New(slog.Default()))
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 30 * time.Second,
	}))

	e.GET("/", handleRoot)
	e.GET("/user", handleListUsers)
	e.GET("/user/:id", handleGetUser)

	return e
}

type userRoute struct {
	ID int `param:"id"`
}

func handleRoot(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{"foo": "bar"})
}

func handleListUsers(c echo.Context) error {
	result := []map[string]any{
		{"user": 111},
		{"user": 222},
	}
	return c.JSON(http.StatusOK, result)
}

func handleGetUser(c echo.Context) error {
	qs := c.QueryParam("xxx")

	var u userRoute
	if err := c.Bind(&u); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"message": "invalid type for user id"})
	}

	slog.InfoContext(c.Request().Context(), "calling get user",
		slog.String("qs", qs),
		slog.Group("user",
			slog.Int("id", u.ID),
		),
	)
	return c.JSON(http.StatusOK, map[string]any{"user": u.ID, "qs": qs})
}
