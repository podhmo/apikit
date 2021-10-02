package design

import "context"

// type HandlerFunc func(ctx context.Context) (interface{}, error)
type Article struct {
	Title string
	Text  string
}

// TODO: pagination

func ListArticle(ctx context.Context) ([]*Article, error) {
	return nil, nil
}
func GetArticle(ctx context.Context, articleId string) (*Article, error) {
	return nil, nil
}

// TODO: regex
