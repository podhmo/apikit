package db

import (
	"m/design"
)

type DB struct {
	Articles map[int64]*design.Article
}
type Config struct {
	URL string `json:"url"`
}

func (c *Config) New() *DB {
	return db
}

var db = &DB{
	Articles: map[int64]*design.Article{
		1: &design.Article{
			ID:    1,
			Title: "foo",
		},
	},
}
