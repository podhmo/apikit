// Code generated by "github.com/podhmo/apikit"; DO NOT EDIT.

package handler

import (
	"context"
	"log"
	"m/foo/action"
	"m/foo/webapi/runtime"
	"net/http"
)

func Hello(getProvider func(*http.Request) (*http.Request, Provider, error)) func(http.ResponseWriter, *http.Request) {
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
		result, err := action.Hello(ctx, logger)
		runtime.HandleResult(w, req, result, err)
	}
}
