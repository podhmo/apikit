package auth

import (
	"m/12with-auth/database"
	"m/12with-auth/design"
	"net/http"

	"github.com/morikuni/failure"
)

type BasicAuthChecker struct {
	Username string
	Password string
}

func (c *BasicAuthChecker) Check(r *http.Request) bool {
	username, password, ok := r.BasicAuth()
	if !ok {
		return false
	}
	return username == c.Username && password == c.Password
}

func LoginRequired(w http.ResponseWriter, req *http.Request) error {
	if !sampleBasicAuth.Check(req) {
		w.Header().Add("WWW-Authenticate", `Basic realm="SECRET AREA"`)
		return failure.NewFailure(design.CodeUnauthorized)
	}
	return nil
}

var sampleBasicAuth = &BasicAuthChecker{
	Username: "this-is-demo",
	Password: "don't-use-this",
}

//----------------------------------------
// with DB
//----------------------------------------

func LoginRequiredWithDB(db *database.DB, w http.ResponseWriter, req *http.Request) error {
	check := func() bool {
		username, password, ok := req.BasicAuth()
		if !ok {
			return false
		}
		u, ok := db.Users[username]
		if !ok {
			return false
		}
		return u.Password == password // dont't use this
	}
	if !check() {
		w.Header().Add("WWW-Authenticate", `Basic realm="SECRET AREA"`)
		return failure.NewFailure(design.CodeUnauthorized)
	}
	return nil
}
