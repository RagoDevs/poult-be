package main

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"slices"

	db "github.com/RagoDevs/poult-be/db/sqlc"
	"github.com/RagoDevs/poult-be/mail"
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
		Name     string `json:"name" validate:"required"`
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
		Name:         input.Name,
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

		dt := mail.MailerData{
			Name:  args.Name,
			Email: args.Email,
			Token: token.Plaintext,
			Url:   app.config.frontend_url,
		}

		if err := app.mailer.SendWelcomeEmail(dt); err != nil {
			slog.Error("error sending welcome email to user", "error", err, "email", args.Email)
		}
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
		Scope:  db.ScopePasswordReset,
		UserID: user.ID,
	}
	err = app.store.DeleteAllToken(c.Request().Context(), d)

	if err != nil {
		slog.Error("error deleting reset password token for user", "error", err)
	}

	app.background(func() {

		dt := mail.MailerData{
			Name:  user.Name,
			Email: user.Email,
			Url:   app.config.frontend_url,
		}

		if err := app.mailer.SendResetCompletedEmail(dt); err != nil {
			slog.Error("error sending reset complete email to user", "error", err, "email", user.Email)
		}
	})

	return c.JSON(http.StatusOK, nil)
}

func (app *application) updateUserNameOrPasswordHandler(c echo.Context) error {

	user := c.Get("user").(*db.GetHashTokenForUserRow)

	var input struct {
		Name            *string `json:"name"`
		CurrentPassword *string `json:"current_password"`
		NewPassword     *string `json:"new_password"`
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, envelope{"error": err.Error()})
	}

	if input.Name == nil && input.CurrentPassword == nil && input.NewPassword == nil {
		return c.JSON(http.StatusBadRequest, envelope{"error": "at least one field is required"})
	}

	if input.NewPassword != nil && len(*input.NewPassword) < 8 {
		return c.JSON(http.StatusBadRequest, envelope{"error": "new password must be at least 8 characters long"})
	}

	u, err := app.store.GetUserByEmail(c.Request().Context(), user.Email)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return c.JSON(http.StatusNotFound, envelope{"error": "user not found"})
		default:
			slog.Error("error fetching user", "error", err)
			return c.JSON(http.StatusInternalServerError, envelope{"error": "internal server error"})
		}
	}

	match, err := db.PasswordMatches(db.Password{
		Plaintext: *input.CurrentPassword,
		Hash:      u.PasswordHash,
	})
	if err != nil {
		slog.Error("error matching password", "error", err)
		return c.JSON(http.StatusInternalServerError, envelope{"error": "internal server error"})
	}

	if !match {
		return c.JSON(http.StatusUnauthorized, envelope{"error": "current password is incorrect"})
	}

	pwd, err := db.SetPassword(*input.NewPassword)
	if err != nil {
		slog.Error("error generating hash password", "error", err)
		return c.JSON(http.StatusInternalServerError, envelope{"error": "internal server error"})
	}

	_, err = app.store.UpdateUser(c.Request().Context(), db.UpdateUserParams{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: pwd.Hash,
		Activated:    u.Activated,
	})

	if err != nil {
		slog.Error("error updating user password", "error", err)
		return c.JSON(http.StatusInternalServerError, envelope{"error": "internal server error"})
	}

	return c.JSON(http.StatusOK, envelope{"message": "password updated successfully"})
}
