package dbconfig

import (
	"m/db"
)

type DBConfig struct {
	Config db.Config `json:"db"`
}

func (c *DBConfig) DB() *db.DB {
	return c.Config.New()
}
