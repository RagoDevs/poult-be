package main

import (
	"log/slog"
	"net/http"

	db "github.com/RagoDevs/poult-be/db/sqlc"
	"github.com/labstack/echo/v4"
)

func (app *application) getChickenHistories(c echo.Context) error {
	ctx := c.Request().Context()
	reasonParam := c.QueryParam("reason")
	typeParam := c.QueryParam("type")

	var chickenHistories []db.ChickenHistory
	var err error

	if reasonParam == "" && typeParam == "" {
		
		chickenHistories, err = app.store.GetAllChickenHistories(ctx)
	} else if reasonParam != "" && typeParam == "" {

		reason := db.ReasonType(reasonParam)
		chickenHistories, err = app.store.GetChickenHistories(ctx, reason)
	} else if reasonParam == "" && typeParam != "" {

		chickenType := db.ChickenType(typeParam)
		chickenHistories, err = app.store.GetChickenHistoriesByType(ctx, chickenType)
	} else {

		reason := db.ReasonType(reasonParam)
		chickenType := db.ChickenType(typeParam)
		chickenHistories, err = app.store.GetChickenHistoriesByTypeAndReason(ctx, db.GetChickenHistoriesByTypeAndReasonParams{
			ChickenType: chickenType,
			Reason:      reason,
		})
	}
	
	if err != nil {
		slog.Error("error fetching chicken history", "error", err)
		return c.JSON(http.StatusInternalServerError, envelope{"error": "internal server error"})
	}

	return c.JSON(http.StatusOK, chickenHistories)
}
