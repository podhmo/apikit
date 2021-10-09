package handler

import (
	"context"
	"m/db"
	"m/design/action"
	"net/http"
	"webruntime"
)

func PostArticleCommentHandler(getProvider func(*http.Request) (*http.Request, Provider, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		req, provider, err := getProvider(req)
		if err != nil {
			webruntime.HandleResult(w, req, nil, err)
			return
		}
		var ctx context.Context = req.Context()

		var pathVars struct {
			ArticleID int64 `path:"articleId,required"`
		}
		{
			if err := webruntime.BindPath(&pathVars, req, "articleId"); err != nil {
				w.WriteHeader(http.StatusNotFound) // todo: some helpers
				webruntime.HandleResult(w, req, nil, err)
				return
			}
		}

		var db *db.DB
		{
			db = provider.DB()
		}

		var input action.PostArticleCommentInput
		{
			if err := webruntime.BindBody(&input, req.Body); err != nil {
				w.WriteHeader(http.StatusBadRequest) // todo: some helpers
				webruntime.HandleResult(w, req, nil, err)
				return
			}
		}
		result, err := action.PostArticleComment(ctx, db, input, pathVars.ArticleID)
		webruntime.HandleResult(w, req, result, err)
	}
}
