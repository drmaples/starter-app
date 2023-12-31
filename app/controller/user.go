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

// @Summary		list all users
// @Description	list all users
// @Tags		users
// @Accept		json
// @Produce		json
// @Security 	ApiKeyAuth
// @Success		200	{object}	[]dto.User
// @Failure		401	{object}	dto.ErrorResponse
// @Failure		500	{object}	dto.ErrorResponse
// @Router		/v1/user [get]
func (con *Controller) handleListUsers(c echo.Context) error {
	ctx := c.Request().Context()

	_, err := con.extractUser(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.NewErrorResp(err.Error()))
	}

	users, err := con.userRepo.ListUsers(ctx, con.db, repo.DefaultSchema)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResp(err.Error()))
	}

	var res dto.User
	return c.JSON(http.StatusOK, res.FromModels(users))
}

// @Summary		get user by id
// @Description	get user by id
// @Tags		users
// @Accept		json
// @Produce		json
// @Security 	ApiKeyAuth
// @Param 		id path int true "user id"
// @Success		200	{object}	dto.User
// @Failure		400	{object}	dto.ErrorResponse
// @Failure		401	{object}	dto.ErrorResponse
// @Failure		404	{object}	dto.ErrorResponse
// @Failure		500	{object}	dto.ErrorResponse
// @Router		/v1/user/{id} [get]
func (con *Controller) handleGetUser(c echo.Context) error {
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

	u, err := con.userRepo.GetUserByID(ctx, con.db, repo.DefaultSchema, ur.ID)
	if err != nil {
		if errors.Is(err, repo.ErrNoRowsFound) {
			return c.JSON(http.StatusNotFound, dto.NewErrorResp("no user for given id"))
		}
	}

	var res dto.User
	return c.JSON(http.StatusOK, res.FromModel(*u))
}

// @Summary		create user
// @Description	create user
// @Tags		users
// @Accept		json
// @Produce		json
// @Security 	ApiKeyAuth
// @Param 		data body dto.CreateUser true "data"
// @Success		200	{object}	dto.User
// @Failure		400	{object}	dto.ErrorResponse
// @Failure		401	{object}	dto.ErrorResponse
// @Failure		404	{object}	dto.ErrorResponse
// @Failure		500	{object}	dto.ErrorResponse
// @Router		/v1/user [post]
func (con *Controller) handleCreateUser(c echo.Context) error {
	ctx := c.Request().Context()

	var u dto.CreateUser
	if err := c.Bind(&u); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResp(err.Error()))
	}
	if err := c.Validate(u); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResp(err.Error()))
	}

	tx, err := con.db.BeginTx(ctx, nil)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResp(err.Error()))
	}

	newUser, err := con.userRepo.CreateUser(ctx, tx, repo.DefaultSchema, u.Model())
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

	var res dto.User
	return c.JSON(http.StatusOK, res.FromModel(*newUser))
}
