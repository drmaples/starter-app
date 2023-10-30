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

	// loggedInUser, err := extractUser(c)
	// if err != nil {
	// 	return c.JSON(http.StatusUnauthorized, dto.NewErrorResp(err.Error()))
	// }
	// fmt.Println(">>>>>>>> logged in user:", loggedInUser)

	users, err := repo.NewUserRepo().ListUsers(ctx, repo.DBConn(), repo.DefaultSchema)
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

	u, err := repo.NewUserRepo().GetUserByID(ctx, repo.DBConn(), repo.DefaultSchema, ur.ID)
	if err != nil {
		if errors.Is(err, repo.ErrNoRowsFound) {
			return c.JSON(http.StatusNotFound, dto.NewErrorResp("no user for given id"))
		}
	}

	return c.JSON(http.StatusOK, u)
}

func handleCreateUser(c echo.Context) error {
	ctx := c.Request().Context()

	var u dto.CreateUser
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

	newUser, err := repo.NewUserRepo().CreateUser(ctx, tx, repo.DefaultSchema, u.Model())
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
