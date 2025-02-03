package response

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

const (
	StatusOK    = "ОК"
	StatusError = "Error"
)

func OK() Response {
	return Response{
		Status: StatusOK,
	}
}

func Error(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
	}
}

func ValidationError(errs validator.ValidationErrors) Response {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is a required field", err.Field()))
		case "url":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not a valid URL", err.Field()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not valid", err.Field()))
		}
	}

	return Response{
		Status: StatusError,
		Error:  strings.Join(errMsgs, ", "),
	}
}

func ValidationErrorRegisterUser(errs validator.ValidationErrors) Response {
	var errMsgs []string

	for _, err := range errs {
		switch err.Tag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("Field '%s' is required", err.Field()))
		case "email":
			errMsgs = append(errMsgs, fmt.Sprintf("Field '%s' must be a valid email address", err.Field()))
		case "min":
			errMsgs = append(errMsgs, fmt.Sprintf("Field '%s' must have at least %s characters", err.Field(), err.Param()))
		case "max":
			errMsgs = append(errMsgs, fmt.Sprintf("Field '%s' can have at most %s characters", err.Field(), err.Param()))
		case "alphanum":
			errMsgs = append(errMsgs, fmt.Sprintf("Field '%s' must contain only alphanumeric characters", err.Field()))
		case "eqfield":
			errMsgs = append(errMsgs, fmt.Sprintf("Field '%s' must be equal to field '%s'", err.Field(), err.Param()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("Field '%s' is invalid", err.Field()))
		}
	}

	return Response{
		Status: StatusError,
		Error:  strings.Join(errMsgs, ", "),
	}
}
