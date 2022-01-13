// Code generated by "github.com/podhmo/apikit"; DO NOT EDIT.

package handler

import (
	"context"
	"m/13openapi/design"
	"m/13openapi/design/enum"
	"m/13openapi/handler/runtime"
	"net/http"
)

func ListArticle(getProvider func(*http.Request) (*http.Request, Provider, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		req, provider, err := getProvider(req)
		if err != nil {
			runtime.HandleResult(w, req, nil, err)
			return
		}
		var ctx context.Context = req.Context()
		var db *design.DB
		{
			var err error
			db, err = provider.DB(ctx)
			if err != nil {
				runtime.HandleResult(w, req, nil, err)
				return
			}
		}
		var queryParams struct {
			limit *int            `query:"limit"`
			sort  *enum.SortOrder `query:"sort"`
		}
		if err := runtime.BindQuery(&queryParams, req); err != nil {
			_ = err // ignored
		}
		result, err := design.ListArticle(ctx, db, queryParams.limit, queryParams.sort)
		runtime.HandleResult(w, req, result, err)
	}
}