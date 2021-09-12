package translate

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/podhmo/apikit/difftest"
	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/tinypkg"
)

type Session struct{}
type Todo struct {
	Title string
	Done  bool
}

func AddTodo(session *Session, title string, done bool) (*Todo, error) {
	return nil, nil
}
func AddTodoWithContext(ctx context.Context, session *Session, title string, done bool) (*Todo, error) {
	return nil, nil
}

func TestWriteRunner(t *testing.T) {
	main := tinypkg.NewPackage("main", "")
	resolver := resolve.NewResolver()

	cases := []struct {
		name  string
		input interface{}
		here  *tinypkg.Package
		want  string

		wantError     error
		modifyTracker func(t *Tracker)
	}{
		{
			name:  "RunAddTodo",
			input: AddTodo,
			here:  main,
			want: `
import (
	"github.com/podhmo/apikit/translate"
	"m/component"
)
func RunAddTodo(provider component.Provider, title string, done bool) (*translate.Todo, error) {
	var session *translate.Session
	{
		session = provider.Session()
	}
	return translate.AddTodo(session, title, done)
}`,
		},
		{
			name:  "RunAddTodoWithContext",
			input: AddTodoWithContext,
			here:  main,
			want: `
import (
	"context"
	"github.com/podhmo/apikit/translate"
	"m/component"
)
func RunAddTodoWithContext(ctx context.Context, provider component.Provider, title string, done bool) (*translate.Todo, error) {
	var session *translate.Session
	{
		session = provider.Session()
	}
	return translate.AddTodoWithContext(ctx, session, title, done)
}`,
		},
		{
			name:  "RunAddTodoWithOverride",
			input: AddTodo,
			here:  main,
			modifyTracker: func(tracker *Tracker) {
				rt := reflect.TypeOf(AddTodo).In(0)
				def := resolver.Def(func() (*Session, error) { return nil, nil })
				tracker.Override(rt, "session", def)
			},
			want: `
import (
	"github.com/podhmo/apikit/translate"
	"m/component"
)
func RunAddTodoWithOverride(provider component.Provider, title string, done bool) (*translate.Todo, error) {
	var session *translate.Session
	{
		var err error
		session, err = provider.Session()
		if err != nil {
			return nil, err
		}
	}
	return translate.AddTodo(session, title, done)
}`,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			translator := NewTranslator(resolver)
			if c.modifyTracker != nil {
				c.modifyTracker(translator.Tracker)
			}

			def := resolver.Def(c.input)

			providerSymbol := tinypkg.NewPackage("m/component", "").NewSymbol("Provider")
			provider := &tinypkg.Var{Name: "provider", Node: providerSymbol}

			code := translator.TranslateToRunner(c.here, def, c.name, provider)
			var buf bytes.Buffer

			if err := code.EmitImports(&buf); err != nil {
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
