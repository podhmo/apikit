package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Mount(r chi.Router, getProvider func(*http.Request) (*http.Request, Provider, error)) {
	r.Post("/articles/{articleId}/comments", PostArticleCommentHandler(getProvider))
}
