//go:build apikit
// +build apikit

package main

import (
	"encoding/json"
	"m/30graphql/action"
	"m/30graphql/design"
	"os"

	"github.com/podhmo/apikit/graph"
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

	r.Object(design.Chat{})
	r.Object(design.User{})

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
	r := router()
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(r)
}
