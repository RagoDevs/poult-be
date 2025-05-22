package main

import (
	"log/slog"
	"net/http"

	db "github.com/RagoDevs/poult-be/db/sqlc"
	"github.com/labstack/echo/v4"
)

func (app *application) addTxnTrackerhandler(c echo.Context) error {

	var input db.TxnRequest

	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, envelope{"error": err.Error()})
	}

	if err := app.validator.Struct(input); err != nil {
		return c.JSON(http.StatusBadRequest, envelope{"error": err.Error()})
	}

	if err := app.store.TxnCreateTransaction(c.Request().Context(), input); err != nil {
		slog.Error("error creating transaction", "error", err)
		return c.JSON(http.StatusInternalServerError, envelope{"error": "internal server error"})
	}

	return c.JSON(http.StatusOK, nil)

}
