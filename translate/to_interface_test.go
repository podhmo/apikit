package translate

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/podhmo/apikit/pkg/difftest"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
)

type DB struct{}
type User struct{}

func ListUser(db *DB) []User                   { return nil }
func ListName(db *DB) []string                 { return nil }
func ListUserWithError(db *DB) ([]User, error) { return nil, nil }

func ListUserFromAnotherDB(anotherDb *DB) []User { return nil }

func TestInterface(t *testing.T) {
	resolver := resolve.NewResolver()

	main := resolver.NewPackage("main", "main")
	pkg := resolver.NewPackage(reflect.TypeOf(DB{}).PkgPath(), "")

	config := DefaultConfig()
	config.Resolver = resolver

	cases := []struct {
		msg   string
		input []interface{}
		here  *tinypkg.Package
		want  string

		wantError     error
		modifyTracker func(t *resolve.Tracker)
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
			msg:   "1 component, N actions, another package",
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
		{
			msg:   "with override, another package",
			here:  main,
			input: []interface{}{ListUser},
			modifyTracker: func(tracker *resolve.Tracker) {
				rt := reflect.TypeOf(ListUser).In(0)
				def := resolver.Def(func() (*DB, error) { return nil, nil })
				tracker.Override(rt, "db", def)
			},
			want: `
import (
	"github.com/podhmo/apikit/translate"
)
type Component interface {
	DB() (*translate.DB, error)
}`,
		},
		{
			msg:   "with override, with context",
			here:  main,
			input: []interface{}{ListUser},
			modifyTracker: func(tracker *resolve.Tracker) {
				rt := reflect.TypeOf(ListUser).In(0)
				def := resolver.Def(func(ctx context.Context) (*DB, error) { return nil, nil })
				tracker.Override(rt, "db", def)
			},
			want: `
import (
	"context"
	"github.com/podhmo/apikit/translate"
)
type Component interface {
	DB(ctx context.Context) (*translate.DB, error)
}`,
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.msg, func(t *testing.T) {
			translator := NewTranslator(config)
			for _, useFn := range c.input {
				def := resolver.Def(useFn)
				translator.Tracker.Track(def)
			}
			if c.modifyTracker != nil {
				c.modifyTracker(translator.Tracker)
			}

			code := translator.TranslateToInterface(c.here, "Component")
			var buf strings.Builder
			collector := tinypkg.NewImportCollector(c.here)
			if err := code.CollectImports(collector); err != nil && c.wantError == nil || c.wantError != err {
				t.Fatalf("unexpected error, collect imports %+v", err)
			}
			imports := collector.Imports
			if err := code.EmitImports(&buf, imports); err != nil {
				if c.wantError == nil || c.wantError != err {
					t.Fatalf("unexpected error, import %+v", err)
				}
			}
			if err := code.EmitCode(&buf); err != nil {
				if c.wantError == nil || c.wantError != err {
					t.Fatalf("unexpected error, code %+v", err)
				}
			}

			got := buf.String()
			if want, got := strings.TrimSpace(c.want), strings.TrimSpace(got); want != got {
				difftest.LogDiffGotStringAndWantString(t, got, want)
			}
		})
	}
}
