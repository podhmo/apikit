package action

import (
	"context"
	"fmt"
	"m/db"
	"m/design"
	"m/util/validate"
)

type PostArticleCommentInput struct {
	Text string `validate:"required"`
}

func PostArticleComment(
	ctx context.Context,
	db *db.DB,
	input PostArticleCommentInput,
	articleID int64,
) (*design.Article, error) {
	if err := validate.Validate(input); err != nil {
		return nil, err // 400 or 422
	}
	article, ok := db.Articles[articleID]
	if !ok {
		return nil, fmt.Errorf("not found") // 404
	}
	article.Comments = append(article.Comments, &design.Comment{
		Author: "someone",
		Text:   input.Text,
	})
	return article, nil
}
