package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"webruntime"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
)

type Article struct {
	ID       int64      `json:"id"`
	Title    string     `json:"title"`
	Text     string     `json"text"`
	Comments []*Comment `json:"comments"`
}
type Comment struct {
	Author string `json:"author"`
	Text   string `json"text"`
}

var articles = map[int64]*Article{
	1: &Article{
		ID:    1,
		Title: "foo",
	},
}

func PostArticleCommentHandler(w http.ResponseWriter, req *http.Request) {
	var pathItem struct {
		ArticleID int64 `path:"articleId,required"`
	}

	{
		if err := webruntime.BindPath(&pathItem, req, "articleId"); err != nil {
			w.WriteHeader(http.StatusNotFound) // todo: some helpers
			webruntime.HandleResult(w, req, nil, err)
			return
		}
	}

	var input PostArticleCommentInput
	{
		if err := webruntime.BindBody(&input, req.Body); err != nil {
			w.WriteHeader(http.StatusBadRequest) // todo: some helpers
			webruntime.HandleResult(w, req, nil, err)
			return
		}
	}
	ctx := req.Context()
	result, err := PostArticleComment(ctx, input, pathItem.ArticleID)
	webruntime.HandleResult(w, req, result, err)
}

type PostArticleCommentInput struct {
	Text string `json:"text" validate:"required"`
}

func PostArticleComment(
	ctx context.Context,
	input PostArticleCommentInput,
	articleID int64,
) (*Article, error) {
	if err := webruntime.Validate(input); err != nil {
		return nil, err
	}
	article, ok := articles[articleID]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	article.Comments = append(article.Comments, &Comment{
		Author: "someone",
		Text:   input.Text,
	})
	return article, nil
}

var validate = validator.New()

func Validate(ob interface{}) error {
	// TODO: merge error
	if err := validate.Struct(ob); err != nil {
		return err
	}
	if v, ok := ob.(interface{ Validate() error }); ok {
		return v.Validate() // TODO: 422
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("!! %+v", err)
	}
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

	r.Post("/articles/{articleId}/comments", PostArticleCommentHandler)

	port := 8888
	if v, err := strconv.Atoi(os.Getenv("PORT")); err == nil {
		port = v
	}
	addr := fmt.Sprintf(":%d", port)
	log.Println("listen ...", addr)
	return http.ListenAndServe(addr, r)
}
