package translate

import (
	"reflect"
	"strings"
	"testing"

	"github.com/podhmo/apikit/difftest"
	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/tinypkg"
)

type DB struct{}
type User struct{}

func ListUser(db *DB) []User                   { return nil }
func ListName(db *DB) []string                 { return nil }
func ListUserWithError(db *DB) ([]User, error) { return nil, nil }

func ListUserFromAnotherDB(anotherDb *DB) []User { return nil }

func TestInterface(t *testing.T) {
	main := tinypkg.NewPackage("main", "main")
	pkg := tinypkg.NewPackage(reflect.TypeOf(DB{}).PkgPath(), "")
	resolver := resolve.NewResolver()

	cases := []struct {
		msg   string
		input []interface{}
		here  *tinypkg.Package
		want  string
	}{
		{
			msg:   "1 component, another package",
			here:  main,
			input: []interface{}{ListUser},
			want: `
import (
	"github.com/podhmo/apikit/translate"
)
type Component interface {
	DB() *translate.DB
}`,
		},
		{
			msg:   "1 component, same package",
			here:  pkg,
			input: []interface{}{ListUser},
			want: `
type Component interface {
	DB() *DB
}`,
		},
		// TODO: support qualified import
		{
			msg:   "N component, another package",
			here:  main,
			input: []interface{}{ListUser, ListName, ListUserWithError},
			want: `
import (
	"github.com/podhmo/apikit/translate"
)
type Component interface {
	DB() *translate.DB
}`,
		},
		{
			msg:   "same types but with another name, another package",
			here:  main,
			input: []interface{}{ListUser, ListUserFromAnotherDB},
			want: `
import (
	"github.com/podhmo/apikit/translate"
)
type Component interface {
	Db() *translate.DB
	AnotherDb() *translate.DB
}`,
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.msg, func(t *testing.T) {
			translator := NewTranslator(resolver, c.input)
			code := translator.TranslateInterface(c.here, "Component")

			var buf strings.Builder
			if err := code.EmitImports(&buf); err != nil {
				t.Fatalf("unexpected error, import %+v", err)
			}
			if err := code.EmitCode(&buf); err != nil {
				t.Fatalf("unexpected error, code %+v", err)
			}

			got := buf.String()
			if want, got := strings.TrimSpace(c.want), strings.TrimSpace(got); want != got {
				difftest.LogDiffGotStringAndWantString(t, got, want)
			}
		})
	}
}
