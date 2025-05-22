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

func (app *application) getTransactionsByTypeHandler(c echo.Context) error {
	transactionType := c.Param("transactionType")

	if transactionType == "" {
		return c.JSON(http.StatusBadRequest, envelope{"error": "transactionType path parameter is required"})
	}

	if transactionType != string(db.TransactionTypeExpense) && transactionType != string(db.TransactionTypeIncome) {
		return c.JSON(http.StatusBadRequest, envelope{"error": "invalid transactionType. Must be 'expense' or 'income'"})
	}

	transactions, err := app.store.GetTransactionsByType(c.Request().Context(), db.TransactionType(transactionType))
	if err != nil {
		slog.Error("error fetching transactions by type", "type", transactionType, "error", err)
		return c.JSON(http.StatusInternalServerError, envelope{"error": "failed to retrieve transactions"})
	}

	return c.JSON(http.StatusOK, transactions)
}
