package utils

import (
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/labstack/echo/v4"
)

func GetFileFromForm(c echo.Context, name, purpose string) (*entities.File, error) {
	file, err := c.FormFile(name)
	if err != nil {
		return nil, err
	}
	body, err := file.Open()
	if err != nil {
		return nil, err
	}
	return &entities.File{
		Name:    file.Filename,
		Body:    body,
		Purpose: purpose,
	}, nil
}
