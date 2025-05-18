package main

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"slices"

	db "github.com/RagoDevs/poult-be/db/sqlc"
	"github.com/labstack/echo/v4"
)

type SignupData struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Token string `json:"token"`
}

type ResetCompleteData struct {
	Email string `json:"email"`
}

func (app *application) registerUserHandler(c echo.Context) error {

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

	emails := strings.Split(app.config.emails, ",")

	found := slices.Contains(emails, input.Email)

	if !found {
		return c.JSON(http.StatusUnauthorized, envelope{"error": "email not allowed"})
	}

	pwd, err := db.SetPassword(input.Password)

	if err != nil {
		slog.Error("error generating hash password", "error", err)
		return err
	}

	args := db.CreateUserParams{
		Email:        input.Email,
		PasswordHash: pwd.Hash,
		Activated:    false,
	}

	a, err := app.store.CreateUser(c.Request().Context(), args)

	if err != nil {
		switch {

		case err.Error() == db.DuplicateEmail:
			return c.JSON(http.StatusBadRequest, envelope{"error": "email is already in use"})

		default:
			slog.Error("error creating user", "error", err)
			return c.JSON(http.StatusInternalServerError, envelope{"error": "internal server error"})
		}

	}

	expiry := time.Now().Add(3 * 24 * time.Hour)

	token, err := app.store.NewToken(a.ID, expiry, db.ScopeActivation)
	if err != nil {
		slog.Error("error generating new token", "error", err)
		return c.JSON(http.StatusInternalServerError, envelope{"error": "internal server error"})
	}

	app.background(func() {

		dt := SignupData{
			ID:    a.ID.String(),
			Email: args.Email,
			Token: token.Plaintext,
		}

		jsonData, err := json.Marshal(dt)
		if err != nil {
			slog.Error("Error marshaling JSON", "Error", err)
		}

		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/signup", app.config.mailer_url), bytes.NewBuffer(jsonData))
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

		slog.Info(fmt.Sprintf("Email sent successfully to %s\n", args.Email))

	})

	return c.JSON(http.StatusCreated, nil)
}

func (app *application) activateUserHandler(c echo.Context) error {

	var input struct {
		TokenPlaintext string `json:"token" validate:"required,len=26"`
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, envelope{"error": err.Error()})
	}

	if err := app.validator.Struct(input); err != nil {
		return c.JSON(http.StatusBadRequest, envelope{"error": err.Error()})
	}

	tokenHash := sha256.Sum256([]byte(input.TokenPlaintext))

	args := db.GetHashTokenForUserParams{
		Scope:  db.ScopeActivation,
		Hash:   tokenHash[:],
		Expiry: time.Now(),
	}

	user, err := app.store.GetHashTokenForUser(c.Request().Context(), args)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			slog.Error("error fetching token user", "error", err)
			return c.JSON(http.StatusNotFound, envelope{"error": "invalid token or expired"})
		default:
			slog.Error("error fetching token user admin", "error", err)
			return c.JSON(http.StatusInternalServerError, envelope{"error": "internal server error"})
		}

	}

	if user.Activated {
		return c.JSON(http.StatusBadRequest, envelope{"error": "user arleady actvated"})
	}

	param := db.UpdateUserParams{

		ID:           user.ID,
		Email:        user.Email,
		Activated:    true,
		PasswordHash: user.PasswordHash,
	}
	_, err = app.store.UpdateUser(c.Request().Context(), param)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			slog.Error("error conflict updating admin ", "error", err)
			return c.JSON(http.StatusConflict, envelope{"error": "unable to complete request due to an edit conflict"})
		default:
			slog.Error("error updating admin ", "error", err)
			return c.JSON(http.StatusInternalServerError, envelope{"error": "internal server error"})
		}

	}

	return c.JSON(http.StatusOK, nil)
}

func (app *application) updateUserPasswordOnResetHandler(c echo.Context) error {

	var input struct {
		Password       string `json:"password" validate:"required,min=8"`
		TokenPlaintext string `json:"token" validate:"required,len=26"`
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, envelope{"error": err.Error()})
	}

	if err := app.validator.Struct(input); err != nil {
		return c.JSON(http.StatusBadRequest, envelope{"error": err.Error()})
	}

	tokenHash := sha256.Sum256([]byte(input.TokenPlaintext))

	args := db.GetHashTokenForUserParams{
		Scope:  db.ScopePasswordReset,
		Hash:   tokenHash[:],
		Expiry: time.Now(),
	}

	user, err := app.store.GetHashTokenForUser(c.Request().Context(), args)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			slog.Error("error fetching token user", "error", err)
			return c.JSON(http.StatusNotFound, envelope{"errors": "invalid token"})
		default:
			slog.Error("error fetching token user admin", "error", err)
			return c.JSON(http.StatusInternalServerError, envelope{"error": "internal server error"})
		}
	}

	pwd, err := db.SetPassword(input.Password)

	if err != nil {
		return err
	}

	_, err = app.store.UpdateUser(c.Request().Context(), db.UpdateUserParams{
		Email:        user.Email,
		PasswordHash: pwd.Hash,
		Activated:    true,
		ID:           user.ID,
	})

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return c.JSON(http.StatusConflict, envelope{"error": "unable to complete request due to an edit conflict"})
		default:
			slog.Error("error updating admin ", "error", err)
			return c.JSON(http.StatusInternalServerError, envelope{"error": "internal server error"})
		}
	}

	d := db.DeleteAllTokenParams{
		Scope: db.ScopePasswordReset,
		UserID:    user.ID,
	}
	err = app.store.DeleteAllToken(c.Request().Context(), d)

	if err != nil {
		slog.Error("error deleting reset password token for user", "error", err)
	}

	app.background(func() {

		dt := ResetCompleteData{
			Email: user.Email,
		}

		jsonData, err := json.Marshal(dt)
		if err != nil {
			slog.Error("Error marshaling JSON", "Error", err)
		}

		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/completedpwdreset", app.config.mailer_url), bytes.NewBuffer(jsonData))
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

		slog.Info(fmt.Sprintf("Email sent successfully to %s", user.Email))

	})

	return c.JSON(http.StatusOK, nil)
}
