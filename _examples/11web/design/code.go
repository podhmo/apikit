package design

import (
	"github.com/morikuni/failure"
)

// error codes for your application.
const (
	NotFound        failure.StringCode = "NotFound"
	Forbidden       failure.StringCode = "Forbidden"
	ValidationError failure.StringCode = "ValidationError"
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
	case NotFound:
		return 404 // http.StatusNotFound
	case Forbidden:
		return 403 // http.StatusForbidden
	case ValidationError:
		return 422 // http.StatusUnprocessableEntity // or http.StatusBadRequest?
	default:
		return 500 // http.StatusInternalServerError
	}
}
