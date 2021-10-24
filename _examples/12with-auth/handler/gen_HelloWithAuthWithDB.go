// Code generated by "github.com/podhmo/apikit"; DO NOT EDIT.

package handler

import (
	"context"
	"log"
	"m/12with-auth/action"
	"m/12with-auth/auth"
	"m/12with-auth/database"
	"m/12with-auth/runtime"
	"net/http"
)

func HelloWithAuthWithDB(getProvider func(*http.Request) (*http.Request, Provider, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		req, provider, err := getProvider(req)
		if err != nil {
			runtime.HandleResult(w, req, nil, err)
			return
		}
		var ctx context.Context = req.Context()
		var logger *log.Logger
		{
			logger = provider.Logger()
		}
		var db *database.DB
		{
			db = provider.DB()
		}
		if err := auth.LoginRequiredWithDB(db, w, req); err != nil {
			runtime.HandleResult(w, req, nil, err)
			return
		}
		result, err := action.Hello(ctx, logger)
		runtime.HandleResult(w, req, result, err)
	}
}
