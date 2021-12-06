//go:build apikit
// +build apikit

package main

import (
	"m/30graphql/action"
)

// https://github.com/graphql/swapi-graphql

type Router interface {
	Query(options ...Option)
	Mutation(options ...Option)

	Object(ob interface{}, options ...Option)
	Implements(ob interface{}) Option
	Field(name string, r Resolver) Option
	Exclude(names ...string) Option

	// TODO
	Union(fn interface{})
	Enum(values ...interface{})
}

type Field struct {
	Name     string
	Resolver interface{}
}

type Option = interface{}
type Resolver = interface{}

// TODO: interface
// TODO: enum

// // resolver
// func ChatActiveUsers(ctx context.Context, chat *design.Chat) ([]*design.User, error) {
// 	return nil, nil
// }

func mount(r Router) {
	// TODO: required

	r.Query(
		r.Field("me", action.Me),
		r.Field("user", action.User),
		r.Field("allUsers", action.AllUsers),
		// r.Field("search", func(term string) ([]SearchResult, error) { return nil, nil }),
		r.Field("myChats", action.MyChats),
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
}

func main() {
}
