package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationResult_Empty(t *testing.T) {
	res := NewResult(nil, nil)
	res.AddError(nil)
	res.AddErrors(FieldErrors{})

	assert.True(t, res.IsValid())
	assert.Len(t, res.Errors, 0)
	assert.Equal(t, "", res.Error())
	assert.True(t, res.Err() == nil)
}

func TestValidationResult_Invalid(t *testing.T) {
	err1 := FieldError{"error 1", "description 1"}
	err2 := FieldError{"error 2", "description 2"}

	res := NewResult(err1, err2)

	assert.False(t, res.IsValid())
	assert.Len(t, res.Errors, 2)
	assert.Equal(t, res.Errors[0], err1)
	assert.Equal(t, res.Errors[1], err2)
}

func TestValidationResult_AddErrors(t *testing.T) {
	err1 := FieldError{"error 1", "description 1"}
	err2 := FieldError{"error 2", "description 2"}
	err3 := FieldErrors{err1, err2}
	res := NewResult()

	res.AddErrors(err1, err2, err3)

	assert.False(t, res.IsValid())
	assert.Len(t, res.Errors, 4)
	assert.Equal(t, res.Errors[0], err1)
	assert.Equal(t, res.Errors[1], err2)
	assert.Equal(t, res.Errors[2], err1)
	assert.Equal(t, res.Errors[3], err2)
}
