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
)

const (
	stateToken       = "put-state-here"
	oauthCallbackURL = "/backend/google/oauth2_callback"
	authContextKey   = "user"
)

const loginHTML = `<!DOCTYPE html>
<html>
<body>
<a href="%s">Login</a>
</body>
</html>`

func (con *Controller) getOauthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     con.cfg.GoogleClientID,
		ClientSecret: con.cfg.GoogleClientSecret,
		RedirectURL:  con.cfg.ServerAddress + oauthCallbackURL,
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

func (con *Controller) extractUser(c echo.Context) (string, error) {
	rawToken := c.Get(authContextKey)
	if rawToken == nil {
		return "", errors.New("jwt missing")
	}
	token, ok := rawToken.(*jwt.Token)
	if !ok {
		return "", errors.New("jwt is incorrect type")
	}
	claims, ok := token.Claims.(*jwtCustomClaims)
	if !ok {
		return "", errors.New("jwt claims are incorrect type")
	}
	return claims.GetSubject()
}

func (con *Controller) handleLogin(c echo.Context) error {
	// https://developers.google.com/identity/openid-connect/openid-connect#access-type-param
	redirectURL := con.getOauthConfig().AuthCodeURL(
		stateToken,
		// oauth2.AccessTypeOffline, // add if a refresh token is needed
	)
	return c.HTML(http.StatusOK, fmt.Sprintf(loginHTML, redirectURL))
}

func (con *Controller) handleOauthCallback(c echo.Context) error {
	ctx := c.Request().Context()
	code := c.QueryParam("code")
	if c.QueryParam("state") != stateToken { // FIXME: validate this with a nonce
		return c.JSON(http.StatusUnauthorized, dto.NewErrorResp("state token does not match"))
	}

	googleToken, err := con.getOauthConfig().Exchange(ctx, code)
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

	email := googleClaims["email"].(string)
	signedToken, err := NewSignedToken(
		con.cfg.JWTSignKey,
		email,
		googleClaims["name"].(string),
		googleClaims["hd"].(string), // hosted domain
		15*time.Minute,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResp(err.Error()))
	}

	slog.InfoContext(ctx, "successful login", slog.String("email", email))

	return c.JSON(http.StatusOK, echo.Map{"token:": signedToken})
}

func newToken(email string, name string, domain string, expires time.Duration) *jwt.Token {
	now := time.Now()
	return jwt.NewWithClaims(jwt.SigningMethodHS256, &jwtCustomClaims{
		Name:   name,
		Domain: domain, // hosted domain
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Subject:   email,
			ExpiresAt: jwt.NewNumericDate(now.Add(expires)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	})
}

// NewSignedToken returns a signed JWT
func NewSignedToken(jwtSignKey string, email string, name string, domain string, expires time.Duration) (string, error) {
	token := newToken(email, name, domain, expires)
	return token.SignedString([]byte(jwtSignKey))
}
