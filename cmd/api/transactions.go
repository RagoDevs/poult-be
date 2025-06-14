package main

import (
	"log/slog"
	"net/http"
	"time"

	db "github.com/RagoDevs/poult-be/db/sqlc"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (app *application) addTxnTrackerhandler(c echo.Context) error {

	var input struct {
		Type        string    `json:"type" validate:"required,oneof=expense income"`
		Category    string    `json:"category" validate:"required,oneof=food medicine chicken_purchase chicken_sale tools other salary egg_sale"`
		Amount      int32     `json:"amount" validate:"required,gt=0"`
		Date        time.Time `json:"date" validate:"required"`
		Description string    `json:"description" validate:"required"`
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, envelope{"error": err.Error()})
	}

	if err := app.validator.Struct(input); err != nil {
		return c.JSON(http.StatusBadRequest, envelope{"error": err.Error()})
	}

	category, err := app.store.GetCategoryByName(c.Request().Context(), input.Category)
	if err != nil {
		slog.Error("error fetching category", "error", err)
		return c.JSON(http.StatusInternalServerError, envelope{"error": "internal server error"})
	}

	if input.Type == "income" {
		if input.Category != "chicken_sale" && input.Category != "egg_sale" {
			return c.JSON(http.StatusBadRequest, envelope{"error": "category must be chicken_sale or egg_sale for income"})
		}
	}

	if input.Type == "expense" {
		if input.Category != "chicken_purchase" && input.Category != "tools" && input.Category != "medicine" && input.Category != "food" && input.Category != "salary" && input.Category != "other" {
			return c.JSON(http.StatusBadRequest, envelope{"error": "category must be chicken_purchase, tools, medicine, food, salary, or other for expense"})
		}
	}

	if err := app.store.CreateTransaction(c.Request().Context(), db.CreateTransactionParams{
		Type:        db.TransactionType(input.Type),
		CategoryID:  category.ID,
		Amount:      input.Amount,
		Date:        input.Date,
		Description: input.Description,
	}); err != nil {
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

	params := db.GetTransactionsByTypeParams{
		Type:         db.TransactionType(transactionType),
		CategoryName: categoryName,
	}

	transactions, err := app.store.GetTransactionsByType(c.Request().Context(), params)
	if err != nil {
		slog.Error("error fetching transactions by type", "type", transactionType, "category", categoryName, "error", err)
		return c.JSON(http.StatusInternalServerError, envelope{"error": "failed to retrieve transactions"})
	}

	var totalSum int64
	if len(transactions) > 0 {
		totalSum = transactions[0].TotalSum
	}

	return c.JSON(http.StatusOK, envelope{"transactions": transactions, "total_sum": totalSum})
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

func (app *application) deleteTransactionHandler(c echo.Context) error {

	idParam := c.Param("id")
	if idParam == "" {
		return c.JSON(http.StatusBadRequest, envelope{"error": "transaction ID is required"})
	}

	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, envelope{"error": "invalid transaction ID format"})
	}

	_, err = app.store.GetTransaction(c.Request().Context(), id)
	if err != nil {
		slog.Error("error fetching transaction", "error", err)
		return c.JSON(http.StatusNotFound, envelope{"error": "transaction not found"})
	}

	err = app.store.DeleteTransaction(c.Request().Context(), id)
	if err != nil {
		slog.Error("error deleting transaction", "error", err)
		return c.JSON(http.StatusInternalServerError, envelope{"error": "failed to delete transaction"})
	}

	return c.JSON(http.StatusOK, envelope{"message": "transaction deleted successfully"})
}

func (app *application) updateTransactionHandler(c echo.Context) error {
	
	idParam := c.Param("id")
	if idParam == "" {
		return c.JSON(http.StatusBadRequest, envelope{"error": "transaction ID is required"})
	}

	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, envelope{"error": "invalid transaction ID format"})
	}

	_, err = app.store.GetTransaction(c.Request().Context(), id)
	if err != nil {
		slog.Error("error fetching transaction", "error", err)
		return c.JSON(http.StatusNotFound, envelope{"error": "transaction not found"})
	}

	var input struct {
		Type        string    `json:"type" validate:"required,oneof=expense income"`
		Category    string    `json:"category" validate:"required,oneof=food medicine chicken_purchase chicken_sale tools other salary egg_sale"`
		Amount      int32     `json:"amount" validate:"required,gt=0"`
		Date        time.Time `json:"date" validate:"required"`
		Description string    `json:"description" validate:"required"`
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, envelope{"error": err.Error()})
	}

	if err := app.validator.Struct(input); err != nil {
		return c.JSON(http.StatusBadRequest, envelope{"error": err.Error()})
	}

	if input.Type == "income" {
		if input.Category != "chicken_sale" && input.Category != "egg_sale" {
			return c.JSON(http.StatusBadRequest, envelope{"error": "category must be chicken_sale or egg_sale for income"})
		}
	}

	if input.Type == "expense" {
		if input.Category != "chicken_purchase" && input.Category != "tools" && input.Category != "medicine" && input.Category != "food" && input.Category != "salary" && input.Category != "other" {
			return c.JSON(http.StatusBadRequest, envelope{"error": "category must be chicken_purchase, tools, medicine, food, salary, or other for expense"})
		}
	}

	category, err := app.store.GetCategoryByName(c.Request().Context(), input.Category)
	if err != nil {
		slog.Error("error fetching category", "error", err)
		return c.JSON(http.StatusInternalServerError, envelope{"error": "internal server error"})
	}

	err = app.store.UpdateTransaction(c.Request().Context(), db.UpdateTransactionParams{
		ID:          id,
		Type:        db.TransactionType(input.Type),
		CategoryID:  category.ID,
		Amount:      input.Amount,
		Date:        input.Date,
		Description: input.Description,
	})

	if err != nil {
		slog.Error("error updating transaction", "error", err)
		return c.JSON(http.StatusInternalServerError, envelope{"error": "failed to update transaction"})
	}

	return c.JSON(http.StatusOK, envelope{"message": "transaction updated successfully"})
}
