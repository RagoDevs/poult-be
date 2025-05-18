package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	db "github.com/RagoDevs/poult-be/db/sqlc"
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
			return c.JSON(http.StatusNotFound, envelope{"error": "invalid email number or password"})
		default:
			slog.Error("error fetching admin by phone number", "error", err)
			return c.JSON(http.StatusInternalServerError, envelope{"error": "internal server error"})
		}
	}

	if !user.Activated {
		return c.JSON(http.StatusBadRequest, envelope{"error": "user not activated"})
	}

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
		return c.JSON(http.StatusUnauthorized, envelope{"error": "invalid phone number or password"})
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

		dt := ResetData{
			Email: user.Email,
			Token: token.Plaintext,
		}

		jsonData, err := json.Marshal(dt)
		if err != nil {
			slog.Error("Error marshaling JSON", "Error", err)
		}

		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/resetpwd", app.config.mailer_url), bytes.NewBuffer(jsonData))
		if err != nil {
			slog.Error("Error creating request", "Error", err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			slog.Error("Error sending request", "Error", err)
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			slog.Error("Error reading response", "Error", err)
		}

		if resp.StatusCode != http.StatusOK {
			slog.Error(fmt.Sprintf("Request failed with status code %d: %s", resp.StatusCode, string(respBody)))
		}

		slog.Info(fmt.Sprintf("Email sent successfully to %s\n", user.Email))

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

		dt := ActivateData{
			Email: user.Email,
			Token: token.Plaintext,
		}

		jsonData, err := json.Marshal(dt)
		if err != nil {
			slog.Error("Error marshaling JSON", "Error", err)
		}

		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/activate", app.config.mailer_url), bytes.NewBuffer(jsonData))
		if err != nil {
			slog.Error("Error creating request", "Error", err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			slog.Error("Error sending request", "Error", err)
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			slog.Error("Error reading response", "Error", err)
		}

		if resp.StatusCode != http.StatusOK {
			slog.Error(fmt.Sprintf("Request failed with status code %d: %s", resp.StatusCode, string(respBody)))
		}

		slog.Info(fmt.Sprintf("Email sent successfully to %s\n", user.Email))

	})

	return c.JSON(http.StatusAccepted, nil)
}
