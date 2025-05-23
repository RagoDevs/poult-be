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
	categoryName := c.QueryParam("category_name")

	if transactionType == "" {
		return c.JSON(http.StatusBadRequest, envelope{"error": "transactionType path parameter is required"})
	}

	if transactionType != string(db.TransactionTypeExpense) && transactionType != string(db.TransactionTypeIncome) {
		return c.JSON(http.StatusBadRequest, envelope{"error": "invalid transactionType. Must be 'expense' or 'income'"})
	}

	// Create params struct for the query
	params := db.GetTransactionsByTypeParams{
		Type: db.TransactionType(transactionType),
	}

	// Only set CategoryName if it was provided in the query
	if categoryName != "" {
		params.CategoryName = categoryName
	}

	transactions, err := app.store.GetTransactionsByType(c.Request().Context(), params)
	if err != nil {
		slog.Error("error fetching transactions by type", "type", transactionType, "category", categoryName, "error", err)
		return c.JSON(http.StatusInternalServerError, envelope{"error": "failed to retrieve transactions"})
	}

	return c.JSON(http.StatusOK, transactions)
}

type financialSummaryResponse struct {
	TotalIncome   int64 `json:"total_income"`
	TotalExpenses int64 `json:"total_expenses"`
	TotalProfit   int64 `json:"total_profit"`
}

func (app *application) getFinancialSummaryHandler(c echo.Context) error {
	ctx := c.Request().Context()

	totalIncome, err := app.store.GetTotalIncome(ctx)
	if err != nil {
		slog.Error("error fetching total income", "error", err)
		return c.JSON(http.StatusInternalServerError, envelope{"error": "failed to retrieve total income"})
	}

	totalExpenses, err := app.store.GetTotalExpenses(ctx)
	if err != nil {
		slog.Error("error fetching total expenses", "error", err)
		return c.JSON(http.StatusInternalServerError, envelope{"error": "failed to retrieve total expenses"})
	}

	totalProfit := totalIncome - totalExpenses

	response := financialSummaryResponse{
		TotalIncome:   totalIncome,
		TotalExpenses: totalExpenses,
		TotalProfit:   totalProfit,
	}

	return c.JSON(http.StatusOK, response)
}
