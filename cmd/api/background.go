package main

import (
	"context"
	"fmt"
	"log/slog"
)

func (app *application) background(fn func()) {

	app.wg.Add(1)

	go func() {

		defer app.wg.Done()

		defer func() {
			if err := recover(); err != nil {
				slog.LogAttrs(context.Background(),
					slog.LevelError,
					"recovered from panic",
					slog.String("error", fmt.Sprintf("%s", err)),
				)
			}
		}()
		fn()
	}()
}
