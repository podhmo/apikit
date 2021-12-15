package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"webruntime"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Article struct {
	ID       int64      `json:"id"`
	Title    string     `json:"title"`
	Text     string     `json"text"`
	Comments []*Comment `json:"comments"`
}
type Comment struct {
	Author string `json:"author"`
	Text   string `json:"text"`
}

var articles = map[int64]*Article{
	1: &Article{
		ID:    1,
		Title: "foo",
	},
}

type PostArticleCommentInput struct {
	Text string `json:"text" validate:"required"`
}

func PostArticleComment(
	ctx context.Context,
	input PostArticleCommentInput,
	articleID int64,
	loud *bool,
) (*Comment, error) {
	article, ok := articles[articleID]
	if !ok {
		return nil, fmt.Errorf("not found")
	}

	text := input.Text
	if loud != nil && *loud {
		text = strings.ToUpper(text)
	}
	comment := &Comment{
		Author: "someone",
		Text:   text,
	}
	article.Comments = append(article.Comments, comment)
	return comment, nil
}

func PostArticleCommentHandler(w http.ResponseWriter, req *http.Request) {
	var data struct {
		ArticleID int64 `path:"articleId,required"`
		Loud      *bool `query:"loud"`
		PostArticleCommentInput
	}

	// path bindings
	if err := webruntime.BindPathParams(&data, req, "articleId"); err != nil {
		w.WriteHeader(http.StatusNotFound) // todo: some helpers
		webruntime.HandleResult(w, req, nil, err)
		return
	}

	// data bindings
	if err := webruntime.BindBody(&data.PostArticleCommentInput, req.Body); err != nil {
		w.WriteHeader(http.StatusBadRequest) // todo: some helpers
		webruntime.HandleResult(w, req, nil, err)
		return
	}
	if err := webruntime.ValidateStruct(data.PostArticleCommentInput); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity) // todo: some helpers
		webruntime.HandleResult(w, req, nil, err)
		return
	}

	// query bindings
	if err := webruntime.BindQuery(&data, req); err != nil {
		_ = err // ignored
	}

	ctx := req.Context()
	result, err := PostArticleComment(ctx, data.PostArticleCommentInput, data.ArticleID, data.Loud)
	webruntime.HandleResult(w, req, result, err)
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("!! %+v", err)
	}
}

func Mount(r chi.Router) {
	r.Post("/articles/{articleId}/comments", PostArticleCommentHandler)
}

func run() error {
	r := chi.NewRouter()

	// TODO: use httplog
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.Heartbeat("/_ping"))

	Mount(r)

	port := 8888
	if v, err := strconv.Atoi(os.Getenv("PORT")); err == nil {
		port = v
	}
	addr := fmt.Sprintf(":%d", port)
	log.Println("listen ...", addr)
	return http.ListenAndServe(addr, r)
}
