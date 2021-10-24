// Code generated by "github.com/podhmo/apikit"; DO NOT EDIT.

package handler

import (
	"context"
	"m/12with-auth/action"
	"m/12with-auth/runtime"
	"net/http"
)

func Hello(getProvider func(*http.Request) (*http.Request, Provider, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		req, _, err := getProvider(req)
		if err != nil {
			runtime.HandleResult(w, req, nil, err)
			return
		}
		var ctx context.Context = req.Context()
		result, err := action.Hello(ctx)
		runtime.HandleResult(w, req, result, err)
	}
}
