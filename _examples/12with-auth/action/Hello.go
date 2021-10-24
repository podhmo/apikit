package action

import (
	"context"
)

type HelloOutput struct {
	Message string `json: "message"`
}

func Hello(ctx context.Context) (*HelloOutput, error) {
	return &HelloOutput{Message: "hello"}, nil
}
