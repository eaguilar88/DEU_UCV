package security

import (
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type validatedRequest struct {
	Name  string `validate:"required"`
	Email string `validate:"omitempty,email"`
}

func TestCustomValidator_Validate(t *testing.T) {
	v := NewCustomValidator()

	t.Run("valid struct passes", func(t *testing.T) {
		assert.NoError(t, v.Validate(validatedRequest{Name: "Grupo", Email: "grupo@example.com"}))
	})

	t.Run("missing required field is rejected", func(t *testing.T) {
		var verrs validator.ValidationErrors
		err := v.Validate(validatedRequest{})
		assert.True(t, errors.As(err, &verrs), "expected ValidationErrors, got %v", err)
	})

	t.Run("invalid format is rejected", func(t *testing.T) {
		assert.Error(t, v.Validate(validatedRequest{Name: "Grupo", Email: "not-an-email"}))
	})

	t.Run("non-struct returns InvalidValidationError", func(t *testing.T) {
		var invalid *validator.InvalidValidationError
		err := v.Validate("not a struct")
		assert.True(t, errors.As(err, &invalid), "expected InvalidValidationError, got %v", err)
	})
}
