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

	e.GET("/ping", app.ping)

	// User Routes

	e.POST("/users", app.registerUserHandler)
	e.PUT("/users/activate", app.activateUserHandler)
	e.POST("/login", app.createAuthenticationTokenHandler)
	e.POST("/tokens/resend/activation", app.resendActivationTokenHandler)

	// password management
	e.POST("/tokens/password/reset", app.createPasswordResetTokenHandler)
	e.PUT("/users/password/reset", app.updateUserPasswordOnResetHandler)

	g := e.Group("/auth")

	g.Use(app.authenticate)

	//User profile route
	g.PUT("/users/profile", app.updateUserNameOrPasswordHandler)
	g.GET("/users/profile", app.getUserProfileHandler)

	//transaction routes
	g.POST("/transactions", app.addTxnTrackerhandler)
	g.GET("/transactions/type/:transactionType", app.getTransactionsByTypeHandler)
	g.GET("/financial-summary", app.getFinancialSummaryHandler)
	g.DELETE("/transactions/:id", app.deleteTransactionHandler)
	g.PUT("/transactions/:id", app.updateTransactionHandler)

	return e

}
