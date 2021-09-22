package design

import "context"

type HandlerFunc func(ctx context.Context) (interface{}, error)

func Hello(ctx context.Context) (string, error) {
	return "hello", nil
}
