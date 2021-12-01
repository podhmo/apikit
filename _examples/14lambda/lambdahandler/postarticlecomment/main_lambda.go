package main

type Event struct {
	ArticleID int64 `json:"articleID"`
	design.Comment
}
type Provider interface {
}

// see: https://docs.aws.amazon.com/lambda/latest/dg/golang-handler.html

func Handle(getProvider func(context.Context) (context.Context, Provider, error)) func(context.Context, Event) (interface{}, error) {
	return func(ctx context.Context, event Event) (interface{}, error) {
		result, err := design.PostArticleComment()
		if err != nil {
			return nil, err
		}
		return result, nil
	}
}
