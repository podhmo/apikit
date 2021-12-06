package resolver

import "github.com/graphql-go/graphql"

func Schema() (graphql.Schema, error) {
	// Args: graphql.FieldConfigArgument{
	// 	"doneOnly": &graphql.ArgumentConfig{
	// 		Type: graphql.Boolean,
	// 	},
	// },
	roleType := graphql.NewEnum(graphql.EnumConfig{
		Name: "Role",
		Values: graphql.EnumValueConfigMap{
			"ADMIN": &graphql.EnumValueConfig{},
			"USER":  &graphql.EnumValueConfig{},
		},
	})

	userType := graphql.NewObject(graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
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
				Type: roleType,
			},
		},
	})

	chatType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Chat",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.NewNonNull(graphql.ID),
			},
			"users": &graphql.Field{
				Type: graphql.NewList(userType),
			},
			// "messages": &graphql.Fiel{
			// 	Type: graphql.NewList(chatMessageType),
			// },
		},
		// Interfaces: (graphql.InterfacesThunk)(func() []*graphql.Interface {
		// 	return []*graphql.Interface{someInterface}
		// }),
	})

	query := graphql.NewObject(graphql.ObjectConfig{
		Name: "query",
		Fields: graphql.Fields{
			"me": &graphql.Field{
				Type:    userType,
				Resolve: ResolveMe,
			},
			"user": &graphql.Field{
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Type:    userType,
				Resolve: ResolveUser,
			},
			"allUsers": &graphql.Field{
				Type:    graphql.NewList(userType),
				Resolve: ResolveAllUsers,
			},
			"myChats": &graphql.Field{
				Type:    graphql.NewList(chatType),
				Resolve: ResolveMyChats,
			},
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{
		Query: query,
	})
}

func ResolveMe(p graphql.ResolveParams) (interface{}, error) {
	return nil, nil
}

func ResolveUser(p graphql.ResolveParams) (interface{}, error) {
	return nil, nil
}

func ResolveAllUsers(p graphql.ResolveParams) (interface{}, error) {
	return nil, nil
}

func ResolveMyChats(p graphql.ResolveParams) (interface{}, error) {
	return nil, nil
}
