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

// TODO: pagination

func ListArticle(ctx context.Context, db *DB) ([]*Article, error) {
	return nil, nil
}
func GetArticle(ctx context.Context, db *DB, articleID string) (*Article, error) {
	return nil, nil
}

type Comment struct {
	ArticleID string
	Author    string
	Title     string
	Text      string
}

func PostArticleComment(ctx context.Context, db *DB, articleID string, data Comment) (*Comment, error) {
	return nil, nil
}

// TODO: regex
