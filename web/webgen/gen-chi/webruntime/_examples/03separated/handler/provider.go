package handler

import "m/db"

type Provider interface {
	DB() *db.DB
}
