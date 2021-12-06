package resolver

import "github.com/graphql-go/graphql"

type Definitions struct {
	dateType *graphql.Scalar
	roleType *graphql.Enum

	userType        *graphql.Object
	chatType        *graphql.Object
	chatMessageType *graphql.Object
}

func Schema() (graphql.Schema, error) {
	var defs Definitions

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
					Type: graphql.NewList(defs.userType),
				},
				"messages": &graphql.Field{
					Type: graphql.NewList(defs.chatMessageType),
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
					Type: defs.userType,
				},
			}
		}),
	})

	query := graphql.NewObject(graphql.ObjectConfig{
		Name: "query",
		Fields: graphql.Fields{
			"me": &graphql.Field{
				Type:    defs.userType,
				Resolve: ResolveMe,
			},
			"user": &graphql.Field{
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Type:    defs.userType,
				Resolve: ResolveUser,
			},
			"allUsers": &graphql.Field{
				Type:    graphql.NewList(defs.userType),
				Resolve: ResolveAllUsers,
			},
			"myChats": &graphql.Field{
				Type:    graphql.NewList(defs.chatType),
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
