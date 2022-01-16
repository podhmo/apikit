package resolver

import (
	"context"
	"m/30graphql/action"
	"m/30graphql/design"
	"m/30graphql/store"

	"github.com/graphql-go/graphql"
)

type Provider interface {
	Store() *store.Store
}

type getProviderFunc = func(ctx context.Context) Provider

func ResolveMe(getProvider getProviderFunc) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		ctx := p.Context
		provider := getProvider(ctx)
		var store *store.Store
		{
			store = provider.Store()
		}
		return action.Me(ctx, store) // TODO: get current user (?)
	}
}

func ResolveUser(getProvider getProviderFunc) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		ctx := p.Context
		provider := getProvider(ctx)
		var store *store.Store
		{
			store = provider.Store()
		}

		var id design.ID
		if v, ok := p.Args["id"].(design.ID); ok {
			id = v
		}
		return action.User(ctx, store, id)
	}
}

func ResolveAllUsers(getProvider getProviderFunc) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		ctx := p.Context
		provider := getProvider(ctx)
		var store *store.Store
		{
			store = provider.Store()
		}
		return action.AllUsers(ctx, store)
	}
}

func ResolveMyChats(getProvider getProviderFunc) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		ctx := p.Context
		provider := getProvider(ctx)
		var store *store.Store
		{
			store = provider.Store()
		}
		return action.MyChats(ctx, store)
	}
}
func ResolveChatToUsers(getProvider getProviderFunc) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		ctx := p.Context
		provider := getProvider(ctx)
		var store *store.Store
		{
			store = provider.Store()
		}
		chat := p.Source.(*design.Chat)
		return action.ChatToUsers(ctx, store, chat)
	}
}
func ResolveChatToMessages(getProvider getProviderFunc) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		ctx := p.Context
		provider := getProvider(ctx)
		var store *store.Store
		{
			store = provider.Store()
		}
		chat := p.Source.(*design.Chat)
		return action.ChatToMessages(ctx, store, chat)
	}
}
func ResolveChatMessageToUser(getProvider getProviderFunc) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		ctx := p.Context
		provider := getProvider(ctx)
		var store *store.Store
		{
			store = provider.Store()
		}
		message := p.Source.(*design.ChatMessage)
		return action.ChatMessageToUser(ctx, store, message)
	}
}
