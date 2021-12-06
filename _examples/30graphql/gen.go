//go:build apikit
// +build apikit

package main

import "context"

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

// types ----------------------------------------

type ID string
type Date string
type Role string // enum (USER, ADMIN)
const (
	RoleAdmin Role = "ADMIN"
	RoleUser  Role = "USER"
)

type Node struct {
	ID ID
}

type User struct {
	ID       ID
	Username string
	Email    string
	Role     Role
}

type Chat struct {
	ID       ID
	Users    []*User
	Messages []*ChatMessage
}

type ChatMessage struct {
	ID      ID
	Content string
	Time    Date
	User    User
}

type SearchResult interface {
	// User, Chat, ChatMessage
}

// resolver

func ChatActiveUsers(ctx context.Context, chat *Chat) ([]*User, error) {
	return nil, nil
}

func mount(r Router) {
	// TODO: required

	r.Query(
		r.Field("me", func() (*User, error) { return nil, nil }),
		r.Field("user", func(id ID) (*User, error) { return nil, nil }),
		r.Field("allUsers", func() ([]*User, error) { return nil, nil }),
		// r.Field("search", func(term string) ([]SearchResult, error) { return nil, nil }),
		r.Field("myChats", func() ([]*Chat, error) { return nil, nil }),
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
	// r.Union(func(*User, *Chat, *ChatMessage) SearchResult { return nil })
}

func main() {
}
