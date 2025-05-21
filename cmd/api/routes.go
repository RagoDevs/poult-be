package main

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (app *application) routes() http.Handler {

	e := echo.New()

	DefaultCORSConfig := middleware.CORSConfig{
		Skipper:      middleware.DefaultSkipper,
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
	}

	config := middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{Rate: 10, Burst: 30, ExpiresIn: 3 * time.Minute},
		),
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			id := ctx.RealIP()
			return id, nil
		},
		ErrorHandler: func(context echo.Context, err error) error {
			return context.JSON(http.StatusForbidden, nil)
		},
		DenyHandler: func(context echo.Context, identifier string, err error) error {
			return context.JSON(http.StatusTooManyRequests, nil)
		},
	}

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.RateLimiterWithConfig(config))
	e.Use(middleware.CORSWithConfig(DefaultCORSConfig))
	e.Use(middleware.BodyLimit("2K"))

	e.GET("/v1/ping", app.ping)

	// User Routes

	e.POST("/v1/users", app.registerUserHandler)
	e.PUT("/v1/users/activate", app.activateUserHandler)
	e.POST("/v1/login", app.createAuthenticationTokenHandler)
	e.POST("/v1/tokens/resend/activation", app.resendActivationTokenHandler)

	// password management
	e.POST("/v1/tokens/password/reset", app.createPasswordResetTokenHandler)
	e.PUT("/v1/users/password/reset", app.updateUserPasswordOnResetHandler)



	g := e.Group("/v1/auth")

	g.Use(app.authenticate)

	g.GET("/chickens", app.getChickens)
	g.PUT("/chickens/:id", app.UpdateChicken)

	return e

}
