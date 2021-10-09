// Code generated by "github.com/podhmo/apikit"; DO NOT EDIT.

package handler

import (
	"context"
	"m/11web/design"
	"m/11web/runtime"
	"net/http"
)

func PostArticleComment(getProvider func(*http.Request) (*http.Request, Provider, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		articleID := runtime.PathParam(req, "articleId")
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
		result, err := design.PostArticleComment(ctx, db, articleID, data)
		runtime.HandleResult(w, req, result, err)
	}
}
