package controller

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/drmaples/starter-app/app/dto"
	"github.com/drmaples/starter-app/app/platform"
)

const (
	stateToken       = "put-state-here"
	oauthCallbackURL = "/backend/google/oauth2_callback"
)

const loginHTML = `<!DOCTYPE html>
<html>
<body>
<a href="%s">Login</a>
</body>
</html>`

func getOauthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     platform.Config().GoogleClientID,
		ClientSecret: platform.Config().GoogleClientSecret,
		RedirectURL:  GetServerAddress() + oauthCallbackURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}
}

type jwtCustomClaims struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
	jwt.RegisteredClaims
}

func extractUser(c echo.Context) (string, error) {
	token, ok := c.Get("user").(*jwt.Token) // by default token is stored under `user` key
	if !ok {
		return "", errors.New("JWT token missing or invalid")
	}
	claims, ok := token.Claims.(*jwtCustomClaims)
	if !ok {
		return "", errors.New("failed to cast claims as jwt.MapClaims")
	}
	return claims.GetSubject()
}

func handleLogin(c echo.Context) error {
	// https://developers.google.com/identity/openid-connect/openid-connect#access-type-param
	redirectURL := getOauthConfig().AuthCodeURL(
		stateToken,
		// oauth2.AccessTypeOffline, // add if a refresh token is needed
	)
	return c.HTML(http.StatusOK, fmt.Sprintf(loginHTML, redirectURL))
}

func handleOauthCallback(c echo.Context) error {
	ctx := c.Request().Context()
	code := c.QueryParam("code")
	if c.QueryParam("state") != stateToken { // FIXME: validate this with a nonce
		return c.JSON(http.StatusUnauthorized, dto.NewErrorResp("state token does not match"))
	}

	googleToken, err := getOauthConfig().Exchange(ctx, code)
	if err != nil {
		err := errors.Wrap(err, "problem exchanging oauth code for token")
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResp(err.Error()))
	}

	rawToken := googleToken.Extra("id_token").(string)
	googleClaims := jwt.MapClaims{}
	// FIXME: should verify this against google JWK
	if _, _, err := jwt.NewParser().ParseUnverified(rawToken, googleClaims); err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResp(err.Error()))
	}
	now := time.Now()
	email := googleClaims["email"].(string)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwtCustomClaims{
			Name:   googleClaims["name"].(string),
			Domain: googleClaims["hd"].(string), // hosted domain
			RegisteredClaims: jwt.RegisteredClaims{
				ID:        uuid.New().String(),
				Subject:   email,
				ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(now),
			},
		},
	)

	signedToken, err := token.SignedString([]byte(platform.Config().JWTSignKey))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResp(err.Error()))
	}

	slog.InfoContext(ctx, "successful login", slog.String("email", email))

	return c.JSON(http.StatusOK, echo.Map{"token:": signedToken})
}
