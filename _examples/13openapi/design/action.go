package design

import (
	"context"
	"m/13openapi/design/enum"
)

type DB struct{}

func NewDB(ctx context.Context) (*DB, error) {
	return nil, nil
}

type Article struct {
	Title string `json:"title" required:"true"`
	Text  string `json:"text" required:"true"`
}

// TODO: pagination

// ListArticle lists articles
func ListArticle(ctx context.Context, db *DB, limit *int, sort *enum.SortOrder) ([]*Article, error) {
	return nil, nil
}
func GetArticle(ctx context.Context, db *DB, articleID int64) (*Article, error) {
	return nil, nil
}

type Comment struct {
	ArticleID string `json:"articleId"`
	Author    string `json:"author"`
	Title     string `json:"title" required:"true"`
	Text      string `json:"text" required:"true"`
}

type PostArticleCommentInput struct {
	Author string `json:"author"`
	Title  string `json:"title" required:"true"`
	Text   string `json:"text" required:"true"`
}

func PostArticleComment(ctx context.Context, db *DB, articleID int64, data PostArticleCommentInput) (*Comment, error) {
	return nil, nil
}

// TODO: regex
