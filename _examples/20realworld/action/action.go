package action

import (
	"context"
	"fmt"
	"time"
)

type Author struct {
	Username  string `json:"username"`
	Bio       string `json:"bio"`
	Image     string `json:"image"`
	Following bool   `json:"following"`
}

type Article struct {
	Slug           string    `json:"slug"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Body           string    `json:"body"`
	TagList        []string  `json:"tagList"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	Favorited      bool      `json:"favorited"`
	FavoritesCount int       `json:"favoritesCount"`
	Author         Author    `json:"author"`
}

type NewArticle struct {
	Title       string   `json:"title" validate:"required"`
	Description string   `json:"description" validate:"required"`
	Body        string   `json:"body" validate:"required"`
	TagList     []string `json:"tagList"`
}

type CreateArticleInput struct {
	Article NewArticle `json:"article" validate:"required"`
}

// CreateArticle : Create an article. Auth is required
func CreateArticle(ctx context.Context, input CreateArticleInput) (*Author, error) {
	return nil, fmt.Errorf("not impelemented yet")
}
