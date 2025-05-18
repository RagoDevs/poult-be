package main

import (
	"errors"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func ReadUUIDParam(c echo.Context) (uuid.UUID, error) {

	id := c.Param("uuid")

	res := isValidUUID(id)
	if !res {
		return uuid.Nil, errors.New("invalid UUId parameter")
	}

	parsedID, err := uuid.Parse(id)

	if err != nil {
		return uuid.Nil, errors.New("parsing uuid failed")
	}
	return parsedID, nil
}

func isValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
