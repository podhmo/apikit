package auth

import (
	"m/12with-auth/design"
	"net/http"

	"github.com/morikuni/failure"
)

func LoginRequired(req *http.Request) error {
	return failure.NewFailure(design.CodeUnauthorized)
}
