//go:build apikit
// +build apikit

package main

import (
	"context"
	"m/30graphql/action"
	"m/30graphql/design"

	"github.com/podhmo/apikit/graph"
	gengraphqlgo "github.com/podhmo/apikit/graph/gen-graphql-go"
	"github.com/podhmo/apikit/pkg/emitgo"
)

// TODO: interface
// TODO: enum

// // resolver
// func ChatActiveUsers(ctx context.Context, chat *design.Chat) ([]*design.User, error) {
// 	return nil, nil
// }

func router() *graph.Router {
	r := graph.NewRouter()

	r.Query(
		r.Field("me", action.Me),
		r.Field("user", action.User),
		r.Field("allUsers", action.AllUsers),
		// r.Field("search", func(term string) ([]SearchResult, error) { return nil, nil }),
		r.Field("myChats", action.MyChats),
	)

	r.Object(design.User{})
	r.Object(design.Chat{},
		r.Field("users", action.ChatToUsers),
		r.Field("messages", action.ChatToMessages),
	)
	r.Object(design.ChatMessage{},
		r.Field("user", action.ChatMessageToUser),
	)

	// TODO: field resolver
	// r.Object(Chat{},
	// 	r.Field("activeUsers", ChatActiveUsers),
	// )

	// // TODO: enum
	// r.Enum(RoleUser, RoleAdmin)

	// // TODO:	implements interface
	// for _, ob := range []interface{}{User{}, Chat{}, ChatMessage{}} {
	// 	r.Object(ob, r.Implements(Node{}))
	// }

	// // TODO: union
	// r.Union(func(*design.User, *design.Chat, *design.ChatMessage) SearchResult { return nil })
	return r
}

func main() {
	emitgo.NewConfigFromRelativePath(action.AllUsers, "../").EmitWith(func(emitter *emitgo.Emitter) error {
		ctx := context.Background()
		r := router()

		c := gengraphqlgo.DefaultConfig()
		c.DisableFormat = true

		g := c.New(emitter)
		return g.Generate(ctx, r)
		// enc := json.NewEncoder(os.Stdout)
		// enc.SetIndent("", "  ")
		// enc.Encode(r)
	})
}
