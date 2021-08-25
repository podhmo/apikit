package resolve

import (
	"context"
	"fmt"
	"testing"
)

type DB struct{}

type User struct{}

func ListUser(db *DB) []*User {
	return nil
}

func ListUserWithContext(ctx context.Context, db *DB) []*User {
	return nil
}
func ListUserWithFunction(getdb func() (*DB, error)) []*User {
	return nil
}
func ListUserWithInterface(stringer fmt.Stringer) []*User {
	return nil
}

type GetUserInput struct {
	UserID string `path:"userID"`
}

func GetUserWithStruct(db *DB, input GetUserInput) *User {
	return nil
}
func GetUserWithPrimitive(db *DB, userID string) *User {
	return nil
}

// TODO: handle returns

func Test(t *testing.T) {
	resolver := NewResolver()
	cases := []struct {
		Node *Node
		Args []Kind
	}{
		{
			// OK: handle params pointer -> component
			Node: resolver.Resolve(ListUser),
			Args: []Kind{KindComponent},
		},
		{
			// OK: handle params context.Context -> ignored
			Node: resolver.Resolve(ListUserWithContext),
			Args: []Kind{KindIgnored, KindComponent},
		},
		{
			// OK: handle params function -> component
			Node: resolver.Resolve(ListUserWithFunction),
			Args: []Kind{KindComponent},
		},
		{
			// OK: handle params interface -> component
			Node: resolver.Resolve(ListUserWithInterface),
			Args: []Kind{KindComponent},
		},
		{
			// OK: handle params struct -> data
			Node: resolver.Resolve(GetUserWithStruct),
			Args: []Kind{KindComponent, KindData},
		},
		{
			// OK: handle params primitive -> primitive
			Node: resolver.Resolve(GetUserWithPrimitive),
			Args: []Kind{KindComponent, KindPrimitive},
		},
	}

	for _, c := range cases {
		c := c
		node := c.Node

		t.Run(node.Name, func(t *testing.T) {
			t.Logf("input    : %+v", node.GoString())
			t.Logf("signature: %+v", node.Shape.GetReflectType())
			t.Logf("want args: %+v", c.Args)
			for i, x := range node.Args {
				wantKind := c.Args[i]
				gotKind := x.Kind
				t.Run(x.name, func(t *testing.T) {
					if wantKind != gotKind {
						t.Errorf("want %s but got %s", wantKind, gotKind)
					}
				})
			}
		})
	}
}
