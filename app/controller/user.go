package controller

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/drmaples/starter-app/app/dto"
	"github.com/drmaples/starter-app/app/repo"
)

type userRoute struct {
	ID int `param:"id"`
}

func handleListUsers(c echo.Context) error {
	ctx := c.Request().Context()
	tx, err := repo.DBConn().BeginTx(ctx, nil)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResp(err.Error()))
	}

	users, err := repo.NewUserRepo().ListUsers(ctx, tx, schema)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResp(err.Error()))
	}

	return c.JSON(http.StatusOK, users)
}

func handleGetUser(c echo.Context) error {
	ctx := c.Request().Context()

	qs := c.QueryParam("xxx")

	var ur userRoute
	if err := c.Bind(&ur); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResp("invalid type for user id"))
	}

	slog.InfoContext(ctx, "calling get user",
		slog.String("qs", qs),
		slog.Group("user",
			slog.Int("id", ur.ID),
		),
	)

	tx, err := repo.DBConn().BeginTx(ctx, nil)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResp(err.Error()))
	}

	u, err := repo.NewUserRepo().GetUserByID(ctx, tx, schema, ur.ID)
	if err != nil {
		if errors.Is(err, repo.ErrNoRowsFound) {
			return c.JSON(http.StatusNotFound, dto.NewErrorResp("no user for given id"))
		}
	}

	return c.JSON(http.StatusOK, u)
}

func handleCreateUser(c echo.Context) error {
	ctx := c.Request().Context()

	var u dto.User
	if err := c.Bind(&u); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResp(err.Error()))
	}
	if err := c.Validate(u); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResp(err.Error()))
	}

	tx, err := repo.DBConn().BeginTx(ctx, nil)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResp(err.Error()))
	}

	newUser, err := repo.NewUserRepo().CreateUser(ctx, tx, schema, u.Model())
	if err != nil {
		if err := tx.Rollback(); err != nil {
			err := errors.Wrap(err, "problem rolling back transaction")
			return c.JSON(http.StatusInternalServerError, dto.NewErrorResp(err.Error()))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResp(err.Error()))
	}
	if err := tx.Commit(); err != nil {
		err := errors.Wrap(err, "problem committing transaction")
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResp(err.Error()))
	}

	slog.InfoContext(ctx, "added new user",
		slog.Group("user",
			slog.Int("id", newUser.ID),
		),
	)

	return c.JSON(http.StatusOK, newUser)
}
