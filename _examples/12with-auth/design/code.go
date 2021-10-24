package design

import (
	"github.com/morikuni/failure"
)

// error codes for your application.
const (
	CodeNotFound        failure.StringCode = "NotFound"
	CodeUnauthorized    failure.StringCode = "Unauthorized"
	CodeForbidden       failure.StringCode = "Forbidden"
	CodeValidationError failure.StringCode = "ValidationError"
)

func HTTPStatusOf(err error) int {
	if err == nil {
		return 200 // http.StatusOK
	}

	c, ok := failure.CodeOf(err)
	if !ok {
		return 500 // http.StatusInternalServerError
	}
	switch c {
	case CodeUnauthorized:
		return 401
	case CodeForbidden:
		return 403 // http.StatusForbidden
	case CodeNotFound:
		return 404 // http.StatusNotFound
	case CodeValidationError:
		return 422 // http.StatusUnprocessableEntity // or http.StatusBadRequest?
	default:
		return 500 // http.StatusInternalServerError
	}
}
