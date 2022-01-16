package resolver_test

import (
	"context"
	"encoding/json"
	"fmt"
	"m/30graphql/resolver"
	"m/30graphql/store"
	"os"
	"testing"

	"github.com/graphql-go/graphql"
)

type provider struct {
	store *store.Store
}

func (p *provider) Store() *store.Store {
	return p.store
}

func TestIt(*testing.T) {
	p := &provider{store: &store.Store{
		Users: []*store.User{
			{ID: "u1", Username: "foo"},
			{ID: "u2", Username: "bar"},
			{ID: "u3", Username: "boo"},
		},
		Chats: []*store.Chat{
			{ID: "c1", Name: "#random"},
			{ID: "c2", Name: "#general"},
		},
		ChatMessages: []*store.ChatMessage{
			{ID: "m1", ChatID: "c1", UserID: "u1", Content: "hello"},
			{ID: "m2", ChatID: "c1", UserID: "u2", Content: "hello.."},
			{ID: "m3", ChatID: "c1", UserID: "u1", Content: "byebye"},
			{ID: "m4", ChatID: "c2", UserID: "u1", Content: "test test"},
		},
		UserToChats: map[store.UserID][]store.ChatID{
			"u1": []store.ChatID{"c1", "c2"},
			"u2": []store.ChatID{"c1"},
			"u3": nil,
		},
		ChatToUsers: map[store.ChatID][]store.UserID{
			"c1": []store.UserID{"u1", "u2"},
			"c2": []store.UserID{"u1"},
		},
	}}

	s, _ := resolver.NewSchema(func(ctx context.Context) resolver.Provider { return p })

	{
		result := graphql.Do(graphql.Params{
			Schema:        s,
			RequestString: `query { allUsers { id, username } }`,
		})

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		fmt.Println(enc.Encode(result))
	}

	{
		result := graphql.Do(graphql.Params{
			Schema:        s,
			RequestString: `query { myChats { id, messages { content, user { username } } } }`,
		})

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		fmt.Println(enc.Encode(result))
	}
}
