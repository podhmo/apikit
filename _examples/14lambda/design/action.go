package design

import (
	"context"
)

type DB struct{}

func NewDB(ctx context.Context) (*DB, error) {
	return nil, nil
}

type Article struct {
	Title string
	Text  string
}

type Comment struct {
	ArticleID string
	Author    string
	Title     string
	Text      string
}

func PostArticleComment(ctx context.Context, db *DB, articleID int64, data Comment) (*Comment, error) {
	return nil, nil
}
