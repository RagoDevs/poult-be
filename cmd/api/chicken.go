package main

import (
	"log/slog"
	"net/http"

	db "github.com/RagoDevs/poult-be/db/sqlc"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (app *application) getChickens(c echo.Context) error {

	chickens, err := app.store.GetChickens(c.Request().Context())

	if err != nil {
		return c.JSON(http.StatusInternalServerError, envelope{"error": "internal server error"})
	}

	return c.JSON(http.StatusOK, chickens)
}

func (app *application) UpdateChicken(c echo.Context) error {

	var input struct {
		Quantity int32 `json:"quantity" validate:"required"`
	}

	id := c.Param("id")

	uuid, err := uuid.Parse(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, envelope{"error": "invalid id"})
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, envelope{"error": err.Error()})
	}

	if err := app.validator.Struct(input); err != nil {
		return c.JSON(http.StatusBadRequest, envelope{"error": err.Error()})
	}

	err = app.store.UpdateChickenById(c.Request().Context(), db.UpdateChickenByIdParams{
		ID:       uuid,
		Quantity: input.Quantity,
	})
	if err != nil {
		slog.Error("error updating chicken", "error", err)
		return c.JSON(http.StatusInternalServerError, envelope{"error": "internal server error"})
	}

	return c.JSON(http.StatusOK, nil)
}
