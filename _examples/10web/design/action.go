package design

import (
	"context"
)

type DB struct{}

// type HandlerFunc func(ctx context.Context) (interface{}, error)
type Article struct {
	Title string
	Text  string
}

// TODO: pagination

func ListArticle(ctx context.Context, db *DB) ([]*Article, error) {
	return nil, nil
}
func GetArticle(ctx context.Context, db *DB, articleID string) (*Article, error) {
	return nil, nil
}

// TODO: regex
