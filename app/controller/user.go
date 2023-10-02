package controller

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	slogecho "github.com/samber/slog-echo"

	"github.com/drmaples/starter-app/app/repo"
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

	e.GET("/", handleRoot)
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

type userRoute struct {
	ID int `param:"id"`
}

func handleRoot(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{"howdy": "there"})
}

func handleListUsers(c echo.Context) error {
	ctx := c.Request().Context()
	tx, err := repo.DBConn().BeginTx(ctx, nil)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{"message": err.Error()})
	}

	users, err := repo.NewUserRepo().ListUsers(ctx, tx, schema)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{"message": err.Error()})
	}

	return c.JSON(http.StatusOK, users)
}

func handleGetUser(c echo.Context) error {
	ctx := c.Request().Context()

	qs := c.QueryParam("xxx")

	var ur userRoute
	if err := c.Bind(&ur); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"message": "invalid type for user id"})
	}

	slog.InfoContext(ctx, "calling get user",
		slog.String("qs", qs),
		slog.Group("user",
			slog.Int("id", ur.ID),
		),
	)

	tx, err := repo.DBConn().BeginTx(ctx, nil)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{"message": err.Error()})
	}

	u, err := repo.NewUserRepo().GetUserByID(ctx, tx, schema, ur.ID)
	if err != nil {
		if errors.Is(err, repo.ErrNoRowsFound) {
			return c.JSON(http.StatusNotFound, map[string]any{"message": "no user for given id"})
		}
	}

	return c.JSON(http.StatusOK, u)
}

func handleCreateUser(c echo.Context) error {
	ctx := c.Request().Context()

	var u repo.User
	if err := c.Bind(&u); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"message": err.Error()})
	}
	if err := c.Validate(u); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"message": err.Error()})
	}

	tx, err := repo.DBConn().BeginTx(ctx, nil)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{"message": err.Error()})
	}

	newUser, err := repo.NewUserRepo().CreateUser(ctx, tx, schema, u)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			err := errors.Wrap(err, "problem rolling back transaction")
			return c.JSON(http.StatusInternalServerError, map[string]any{"message": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]any{"message": err.Error()})
	}
	if err := tx.Commit(); err != nil {
		err := errors.Wrap(err, "problem committing transaction")
		return c.JSON(http.StatusInternalServerError, map[string]any{"message": err.Error()})
	}

	slog.InfoContext(ctx, "added new user",
		slog.Group("user",
			slog.Int("id", newUser.ID),
		),
	)

	return c.JSON(http.StatusOK, newUser)
}
