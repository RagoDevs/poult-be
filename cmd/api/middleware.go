package main

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	db "github.com/RagoDevs/poult-be/db/sqlc"
	"github.com/labstack/echo/v4"
)

func (app *application) authenticate(next echo.HandlerFunc) echo.HandlerFunc {

	return func(c echo.Context) error {

		authorizationHeader := c.Request().Header.Get("Authorization")

		if authorizationHeader == "" {
			return c.JSON(http.StatusBadRequest, envelope{"error": "missing authentication token"})
		}

		headerParts := strings.Split(authorizationHeader, " ")

		if len(headerParts) != 2 || headerParts[0] != "Bearer" {

			return c.JSON(http.StatusBadRequest, envelope{"error": "invalid token"})

		}

		token := headerParts[1]

		if Isvalid, err := db.IsValidTokenPlaintext(token); !Isvalid {
			return c.JSON(http.StatusBadRequest, envelope{"error": err})
		}

		tokenHash := sha256.Sum256([]byte(token))

		args := db.GetHashTokenForUserParams{
			Scope:  db.ScopeAuthentication,
			Hash:   tokenHash[:],
			Expiry: time.Now(),
		}

		user, err := app.store.GetHashTokenForUser(c.Request().Context(), args)

		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				slog.Error("error", "error", err)
				return c.JSON(http.StatusNotFound, envelope{"error": "invalid token"})
			default:
				slog.Error("error", "error", err)
				return c.JSON(http.StatusInternalServerError, envelope{"error": "internal server error"})
			}
		}

		c.Set("user", user)

		return next(c)

	}

}


