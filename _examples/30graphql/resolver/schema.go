package resolver

import "github.com/graphql-go/graphql"

type Definitions struct {
	dateType *graphql.Scalar
	roleType *graphql.Enum

	userType        *graphql.Object
	chatType        *graphql.Object
	chatMessageType *graphql.Object

	getProvider getProviderFunc
}

func NewDefinitions(getProvider getProviderFunc) *Definitions {
	var defs Definitions
	defs.getProvider = getProvider

	// Args: graphql.FieldConfigArgument{
	// 	"doneOnly": &graphql.ArgumentConfig{
	// 		Type: graphql.Boolean,
	// 	},
	// },
	defs.roleType = graphql.NewEnum(graphql.EnumConfig{
		Name: "Role",
		Values: graphql.EnumValueConfigMap{
			"ADMIN": &graphql.EnumValueConfig{},
			"USER":  &graphql.EnumValueConfig{},
		},
	})

	defs.dateType = graphql.DateTime
	defs.userType = graphql.NewObject(graphql.ObjectConfig{
		Name: "User",
		Fields: (graphql.FieldsThunk)(func() graphql.Fields {
			return graphql.Fields{
				"id": &graphql.Field{
					Type: graphql.NewNonNull(graphql.ID),
				},
				"username": &graphql.Field{
					Type: graphql.String,
				},
				"email": &graphql.Field{
					Type: graphql.String,
				},
				"role": &graphql.Field{
					Type: defs.roleType,
				},
			}
		}),
	})

	defs.chatType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Chat",
		Fields: (graphql.FieldsThunk)(func() graphql.Fields {
			return graphql.Fields{
				"id": &graphql.Field{
					Type: graphql.NewNonNull(graphql.ID),
				},
				"users": &graphql.Field{
					Type:    graphql.NewList(defs.userType),
					Resolve: ResolveChatToUsers(getProvider),
				},
				"messages": &graphql.Field{
					Type:    graphql.NewList(defs.chatMessageType),
					Resolve: ResolveChatToMessages(getProvider),
				},
			}
		}),
		// Interfaces: (graphql.InterfacesThunk)(func() []*graphql.Interface {
		// 	return []*graphql.Interface{someInterface}
		// }),
	})

	defs.chatMessageType = graphql.NewObject(graphql.ObjectConfig{
		Name: "chatMessage",
		Fields: (graphql.FieldsThunk)(func() graphql.Fields {
			return graphql.Fields{
				"id": &graphql.Field{
					Type: graphql.NewNonNull(graphql.ID),
				},
				"content": &graphql.Field{
					Type: graphql.String,
				},
				"time": &graphql.Field{
					Type: defs.dateType,
				},
				"user": &graphql.Field{
					Type:    defs.userType,
					Resolve: ResolveChatMessageToUser(getProvider),
				},
			}
		}),
	})
	return &defs
}

func NewQuery(defs *Definitions) *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "query",
		Fields: graphql.Fields{
			"me": &graphql.Field{
				Type:    defs.userType,
				Resolve: ResolveMe(defs.getProvider),
			},
			"user": &graphql.Field{
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Type:    defs.userType,
				Resolve: ResolveUser(defs.getProvider),
			},
			"allUsers": &graphql.Field{
				Type:    graphql.NewList(defs.userType),
				Resolve: ResolveAllUsers(defs.getProvider),
			},
			"myChats": &graphql.Field{
				Type:    graphql.NewList(defs.chatType),
				Resolve: ResolveMyChats(defs.getProvider),
			},
		},
	})
}

func NewMutation(defs *Definitions) *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
	})
}

func NewSchema(getProvider getProviderFunc) (graphql.Schema, error) {
	defs := NewDefinitions(getProvider)
	return graphql.NewSchema(graphql.SchemaConfig{
		Query:    NewQuery(defs),
		Mutation: NewMutation(defs),
	})
}

/*
// ResolveParams Params for FieldResolveFn()
type ResolveParams struct {
	// Source is the source value
	Source interface{}

	// Args is a map of arguments for current GraphQL request
	Args map[string]interface{}

	// Info is a collection of information about the current execution state.
	Info ResolveInfo

	// Context argument is a context value that is provided to every resolve function within an execution.
	// It is commonly
	// used to represent an authenticated user, or request-specific caches.
	Context context.Context
}

type ResolveInfo struct {
	FieldName      string
	FieldASTs      []*ast.Field
	Path           *ResponsePath
	ReturnType     Output
	ParentType     Composite
	Schema         Schema
	Fragments      map[string]ast.Definition
	RootValue      interface{}
	Operation      ast.Definition
	VariableValues map[string]interface{}
}
*/
