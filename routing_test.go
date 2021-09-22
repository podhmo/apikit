package apikit

import "testing"

// see:  https://github.com/go-chi/chi
// TODO: generics

func TestRouting(t *testing.T) {
	r := NewRouter()
	// TODO: pagination, r.With(paginate).Get("/", ListArticles)

	r.Method("POST", "/", nil)      // createArticle
	r.Method("GET", "/search", nil) // searchArticle

	// TODO: extract path item
	r.Method("GET", "/{articleSlug:[a-z-]+}", nil) // getArticleSlug

	r.Group("/{articleID}", func(r *Router) {
		// r.Use(ArticleCtx)
		r.Get("/", nil)    // GET /articles/123
		r.Put("/", nil)    // PUT /articles/123
		r.Delete("/", nil) // DELETE /articles/123
	})
}
