package grpc_test

import (
	"errors"
	"testing"

	"github.com/nielskrijger/goboot/grpc"
	"github.com/nielskrijger/goboot/validate"
	"github.com/stretchr/testify/assert"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestInternalError_Success(t *testing.T) {
	err := grpc.InternalError

	r := status.Convert(err)
	assert.Equal(t, codes.Internal, r.Code())
	assert.Contains(t, r.Message(), "something went wrong")
}

func TestValidationResult_Nil(t *testing.T) {
	err := grpc.ValidationErrors(nil)

	assert.Nil(t, err)
}

func TestValidationResult_Empty(t *testing.T) {
	err := grpc.ValidationErrors(validate.FieldErrors{})

	assert.Nil(t, err)
}

func TestValidationResult_InvalidError(t *testing.T) {
	err := grpc.ValidationErrors(errors.New("random error"))

	assert.NotNil(t, err)
	r := status.Convert(err)
	assert.Equal(t, "random error", r.Message())
	assert.Equal(t, codes.Internal, r.Code())
}

func TestValidationResult_SingleFieldErrors(t *testing.T) {
	err := grpc.ValidationErrors(validate.FieldErrors{
		{Field: "A", Description: "Message A"},
	})

	assert.NotNil(t, err)
	r := status.Convert(err)
	assert.Equal(t, "field is invalid: A", r.Message())
	assert.Equal(t, codes.InvalidArgument, r.Code())
	detail := r.Details()[0].(*errdetails.BadRequest)
	assert.Equal(t, "A", detail.FieldViolations[0].Field)
	assert.Equal(t, "Message A", detail.FieldViolations[0].Description)
}

func TestValidationResult_MultipleFieldErrors(t *testing.T) {
	err := grpc.ValidationErrors(validate.FieldErrors{
		{Field: "A", Description: "Message A"},
		{Field: "B", Description: "Message B"},
	})

	assert.NotNil(t, err)
	r := status.Convert(err)
	assert.Equal(t, "fields are invalid: A, B", r.Message())
	assert.Equal(t, codes.InvalidArgument, r.Code())
	detail := r.Details()[0].(*errdetails.BadRequest)
	assert.Equal(t, "A", detail.FieldViolations[0].Field)
	assert.Equal(t, "Message A", detail.FieldViolations[0].Description)
	assert.Equal(t, "B", detail.FieldViolations[1].Field)
	assert.Equal(t, "Message B", detail.FieldViolations[1].Description)
}

func TestValidationError_Empty(t *testing.T) {
	err := grpc.ValidationError(nil)

	assert.Nil(t, err)
}

func TestValidationError_InvalidError(t *testing.T) {
	err := grpc.ValidationError(errors.New("random error"))

	assert.NotNil(t, err)
	r := status.Convert(err)
	assert.Equal(t, "random error", r.Message())
	assert.Equal(t, codes.Internal, r.Code())
}

func TestValidationError_Success(t *testing.T) {
	err := grpc.ValidationError(validate.FieldError{Field: "A", Description: "Message A"})

	assert.NotNil(t, err)
	r := status.Convert(err)
	assert.Equal(t, "field is invalid: A", r.Message())
	assert.Equal(t, codes.InvalidArgument, r.Code())
	detail := r.Details()[0].(*errdetails.BadRequest)
	assert.Equal(t, "A", detail.FieldViolations[0].Field)
	assert.Equal(t, "Message A", detail.FieldViolations[0].Description)
}
