package main

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"time"

	db "github.com/RagoDevs/poult-be/db/sqlc"
	"github.com/RagoDevs/poult-be/mail"
	"github.com/labstack/echo/v4"
)

type ActivateData struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

type ResetData struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

func (app *application) createAuthenticationTokenHandler(c echo.Context) error {

	var input struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, envelope{"error": err.Error()})
	}

	if err := app.validator.Struct(input); err != nil {
		return c.JSON(http.StatusBadRequest, envelope{"error": err.Error()})
	}

	user, err := app.store.GetUserByEmail(c.Request().Context(), input.Email)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			slog.Error("error fetching user by email", "error", err)
			return c.JSON(http.StatusNotFound, envelope{"error": "invalid email or password"})
		default:
			slog.Error("error fetching admin by phone number", "error", err)
			return c.JSON(http.StatusInternalServerError, envelope{"error": "internal server error"})
		}
	}

	if !user.Activated {
		return c.JSON(http.StatusBadRequest, envelope{"error": "user not activated"})
	}

	// TO DO
	// send activation email if user not activated

	pwd := db.Password{
		Hash:      user.PasswordHash,
		Plaintext: input.Password,
	}

	match, err := db.PasswordMatches(pwd)

	if err != nil {
		slog.Error("error matching password", "error", err)
		return c.JSON(http.StatusInternalServerError, envelope{"error": "internal server error"})

	}

	if !match {
		slog.Error("error matching password", "error", err)
		return c.JSON(http.StatusUnauthorized, envelope{"error": "invalid email or password"})
	}

	expiry := time.Now().Add(3 * 24 * time.Hour)
	token, err := app.store.NewToken(user.ID, expiry, db.ScopeAuthentication)
	if err != nil {
		slog.Error("error generating new token", "error", err)
		return c.JSON(http.StatusInternalServerError, envelope{"error": "internal server error"})
	}

	return c.JSON(http.StatusCreated, envelope{"token": token.Plaintext, "expiry": expiry.Unix()})
}

func (app *application) createPasswordResetTokenHandler(c echo.Context) error {

	var input struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, envelope{"error": err.Error()})
	}

	if err := app.validator.Struct(input); err != nil {
		return c.JSON(http.StatusBadRequest, envelope{"error": err.Error()})
	}

	user, err := app.store.GetUserByEmail(c.Request().Context(), input.Email)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			slog.Error("error fetching user by email :createPasswordResetTokenHandler", "error", err)
			return c.JSON(http.StatusNotFound, envelope{"error": "user admin not found"})
		default:
			slog.Error("error fetching admin by email :createPasswordResetTokenHandler", "error", err)
			return c.JSON(http.StatusInternalServerError, envelope{"error": "internal server error"})
		}
	}

	if !user.Activated {
		return c.JSON(http.StatusForbidden, envelope{"errors": "account not activated"})
	}

	expiry := time.Now().Add(45 * time.Minute)

	token, err := app.store.NewToken(user.ID, expiry, db.ScopePasswordReset)
	if err != nil {
		slog.Error("error generating token :createPasswordResetTokenHandler", "error", err)
		return c.JSON(http.StatusInternalServerError, envelope{"error": "internal server error"})
	}

	app.background(func() {

		dt := mail.MailerData{
			Name:  user.Name,
			Email: user.Email,
			Token: token.Plaintext,
			Url: app.config.frontend_url,
		}

		if err := app.mailer.SendPasswordResetEmail(dt); err != nil {
			slog.Error("error sending password reset email to user", "error", err, "email", user.Email)
		}

	})

	return c.JSON(http.StatusOK, nil)
}

func (app *application) resendActivationTokenHandler(c echo.Context) error {

	var input struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, envelope{"error": err.Error()})
	}

	if err := app.validator.Struct(input); err != nil {
		return c.JSON(http.StatusBadRequest, envelope{"error": err.Error()})
	}

	user, err := app.store.GetUserByEmail(c.Request().Context(), input.Email)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			slog.Error("error fetching user by email", "error", err)
			return c.JSON(http.StatusNotFound, envelope{"error": "not found"})
		default:
			slog.Error("error fetching admin by email", "error", err)
			return c.JSON(http.StatusInternalServerError, envelope{"error": "internal server error"})
		}
	}

	if user.Activated {
		return c.JSON(http.StatusForbidden, envelope{"errors": "account already activated"})
	}

	expiry := time.Now().Add(3 * 24 * time.Hour)

	token, err := app.store.NewToken(user.ID, expiry, db.ScopeActivation)
	if err != nil {
		slog.Error("error generating new token", "error", err)
		return c.JSON(http.StatusInternalServerError, envelope{"error": "internal server error"})
	}

	app.background(func() {

		dt := mail.MailerData{
			Name: user.Name,
			Email: user.Email,
			Token: token.Plaintext,
			Url: app.config.frontend_url,
		}

		if err := app.mailer.SendActivateEmail(dt); err != nil {
			slog.Error("error sending reactivate email to user", "error", err, "email", user.Email)
		}
	})

	return c.JSON(http.StatusAccepted, nil)
}
