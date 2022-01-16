package postarticlecomment

import (
	"m/14lambda/design"
)

type Event struct {
	ArticleID int64 `json:"articleID"`
	design.Comment
}
