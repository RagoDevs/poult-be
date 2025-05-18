package main

import (
	"net/http"
	"github.com/labstack/echo/v4"
)

func (app *application) ping(c echo.Context) error {

	env := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     app.config.version,
		},
	}
	return c.JSON(http.StatusOK, env)

}