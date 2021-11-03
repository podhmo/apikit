package postarticlecomment

import (
	"context"
	"m/14lambda/design"
	"net/http"
)

func Handle(getProvider func(*http.Request) (*http.Request, Provider, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		req, provider, err := getProvider(req)
		if err != nil {
			HandleResult(w, req, nil, err)
			return
		}
		var ctx context.Context = req.Context()
		var db *design.DB
		{
			db = provider.DB()
		}
		var data design.Comment
		if err := BindBody(&data, req.Body); err != nil {
			w.WriteHeader(400)
			HandleResult(w, req, nil, err)
			return
		}
		if err := ValidateStruct(&data); err != nil {
			w.WriteHeader(422)
			HandleResult(w, req, nil, err)
			return
		}
		result, err := design.PostArticleComment(ctx, db, pathParams.articleID, data)
		HandleResult(w, req, result, err)
	}
}
