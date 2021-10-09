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

type DB struct {
	Articles map[int64]*Article
}
type DBConfigInner struct {
	URL string `json:"url"`
}

var db = &DB{
	Articles: map[int64]*Article{
		1: &Article{
			ID:    1,
			Title: "foo",
		},
	},
}

type Provider interface {
	DB() *DB
}

type PostArticleCommentInput struct {
	Text string `validate:"required"`
}

func PostArticleComment(
	ctx context.Context,
	db *DB,
	input PostArticleCommentInput,
	articleID int64,
) (*Article, error) {
	if err := Validate(input); err != nil {
		return nil, err // 400 or 422
	}
	article, ok := db.Articles[articleID]
	if !ok {
		return nil, fmt.Errorf("not found") // 404
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

		var db *DB
		{
			db = provider.DB()
		}

		var input PostArticleCommentInput
		{
			if err := webruntime.BindBody(&input, req.Body); err != nil {
				w.WriteHeader(http.StatusBadRequest) // todo: some helpers
				webruntime.HandleResult(w, req, nil, err)
				return
			}
		}
		result, err := PostArticleComment(ctx, db, input, pathVars.ArticleID)
		webruntime.HandleResult(w, req, result, err)
	}
}

type DBConfig struct {
	Config DBConfigInner `json:"db"`
}

func (c *DBConfig) DB() *DB {
	return db
}

type Config struct {
	*DBConfig
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

	config := &Config{}
	getProvider := func(req *http.Request) (*http.Request, Provider, error) {
		return req, config, nil
	}
	r.Post("/articles/{articleId}/comments", PostArticleCommentHandler(getProvider))

	port := 8888
	if v, err := strconv.Atoi(os.Getenv("PORT")); err == nil {
		port = v
	}
	addr := fmt.Sprintf(":%d", port)
	log.Println("listen ...", addr)
	return http.ListenAndServe(addr, r)
}
