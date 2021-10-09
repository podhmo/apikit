package resolve

import (
	"context"
	"fmt"
)

// fixtures
type DB struct{}

type User struct{}

func ListUser(db *DB) []*User {
	return nil
}

func ListUserWithPrimitivePointer(db *DB, sort *int, orderBy *string) []*User {
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
func GetUserWithStructWithError(db *DB, input GetUserInput) (*User, error) {
	return nil, nil
}

// https://docs.aws.amazon.com/ja_jp/lambda/latest/dg/golang-handler.html
// func ()
// func () error
// func (TIn), error
// func () (TOut, error)
// func (context.Context) error
// func (context.Context, TIn) error
// func (context.Context) (TOut, error)
// func (context.Context, TIn) (TOut, error)
