// Code generated by "github.com/podhmo/apikit"; DO NOT EDIT.

package handler

import (
	"context"
	"m/12openapi/design"
)

type Provider interface {
	DB(ctx context.Context) (*design.DB, error)
}
