package controller

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	slogecho "github.com/samber/slog-echo"
	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/drmaples/starter-app/app/platform"
	"github.com/drmaples/starter-app/app/repo"
	"github.com/drmaples/starter-app/docs" // docs generated by swag cli
)

// @title Sample App
// @version 1.0
// @description	some description here

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name x-jwt

// Controller contains all info about a controller
type Controller struct {
	e        *echo.Echo
	userRepo repo.IUserRepo
	db       *sql.DB
	cfg      platform.Config
}

// New sets up a new controller
func New(db *sql.DB, cfg platform.Config, userRepo repo.IUserRepo) *Controller {
	e := echo.New()
	con := &Controller{
		e:        e,
		userRepo: userRepo,
		db:       db,
		cfg:      cfg,
	}

	con.adjustDynamicSwaggerInfo()
	con.setupRoutes()

	e.HideBanner = true
	e.HidePort = true
	e.Validator = newValidator()
	e.Use(slogecho.New(slog.Default()))
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 30 * time.Second,
	}))

	// list routes in use like gin. keep at bottom
	if cfg.Environment == "dev" {
		allRoutes := e.Routes()
		sort.SliceStable(allRoutes, func(i, j int) bool {
			return allRoutes[i].Path < allRoutes[j].Path
		})
		for _, r := range allRoutes {
			fmt.Printf("[%-5s] %-35s --> %s\n", strings.ToUpper(r.Method), r.Path, r.Name)
		}
	}

	return con
}

// Run the web server
func (con *Controller) Run(ctx context.Context) {
	slog.InfoContext(ctx, "starting server",
		slog.String("env", con.cfg.Environment),
		slog.String("address", con.cfg.ServerAddress),
	)
	bindAddress := fmt.Sprintf(":%d", con.cfg.ServerPort)
	con.e.Logger.Fatal(con.e.Start(bindAddress))
}

// programmatically set swagger info that changes depending on environment
func (con *Controller) adjustDynamicSwaggerInfo() {
	docs.SwaggerInfo.Host = strings.SplitAfter(con.cfg.ServerAddress, "://")[1] // swagger does not want protocol, builds url dynamically with .Schemes
	docs.SwaggerInfo.Schemes = []string{"http"}
}

func (con *Controller) setupRoutes() {
	unrestricted := con.e.Group("")
	{
		unrestricted.GET("/", func(c echo.Context) error {
			return c.JSON(http.StatusOK, map[string]any{"howdy": "there"})
		})
		unrestricted.GET("/favicon.ico", func(c echo.Context) error { return nil }) // avoids 404 errors in the browser
		unrestricted.GET("/login", con.handleLogin)
		unrestricted.GET(oauthCallbackURL, con.handleOauthCallback)

		unrestricted.GET("/swagger/*", echoSwagger.WrapHandler)
		unrestricted.GET("/docs", func(c echo.Context) error {
			return c.Redirect(http.StatusTemporaryRedirect, "/swagger/index.html")
		})
	}

	restricted := con.e.Group("/v1")
	{
		restricted.Use(
			echojwt.WithConfig(echojwt.Config{
				ContextKey: authContextKey,
				SigningKey: []byte(con.cfg.JWTSignKey),
				NewClaimsFunc: func(c echo.Context) jwt.Claims {
					return new(jwtCustomClaims)
				},
				TokenLookup: "header:x-jwt",
			}),
		)
		restricted.GET("/user", con.handleListUsers)
		restricted.GET("/user/:id", con.handleGetUser)
		restricted.POST("/user", con.handleCreateUser)
	}
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
