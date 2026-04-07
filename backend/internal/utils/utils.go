package utils

import (
	"path/filepath"
	"strings"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/labstack/echo/v4"
)

func GetFileFrom(c echo.Context, key string) (*entities.File, error) {
	file, err := c.FormFile(key)
	if err != nil {
		return nil, err
	}
	body, err := file.Open()
	if err != nil {
		return nil, err
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	return &entities.File{
		Name:    key + ext,
		Body:    body,
		Purpose: key,
	}, nil
}
