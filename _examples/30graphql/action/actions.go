package action

import (
	"context"
	"m/30graphql/design"
)

func ChatActiveUsers(ctx context.Context, chat *design.Chat) ([]*design.User, error) {
	return nil, nil
}

func Me(ctx context.Context) (*design.User, error)                 { return nil, nil }
func User(ctx context.Context, id design.ID) (*design.User, error) { return nil, nil }
func AllUsers(ctx context.Context) ([]*design.User, error)         { return nil, nil }
func MyChats(ctx context.Context) ([]*design.Chat, error)          { return nil, nil }

// r.Field("search", func(term string) ([]SearchResult, error) { return nil, nil }),
