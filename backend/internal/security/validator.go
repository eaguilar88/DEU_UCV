package security

import (
	"github.com/go-playground/validator/v10"
)

type CustomValidator struct {
	validator *validator.Validate
}

func NewCustomValidator() *CustomValidator {
	return &CustomValidator{
		validator: validator.New(),
	}
}

func (cv *CustomValidator) Validate(req any) error {
	if err := cv.validator.Struct(req); err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return err
		}
	}
	return nil
}
