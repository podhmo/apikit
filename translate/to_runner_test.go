package translate

import (
	"bytes"
	"context"
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
		name      string
		input     interface{}
		here      *tinypkg.Package
		want      string
		wantError error
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
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			translator := NewTranslator(resolver)
			def := resolver.Def(c.input)

			providerSymbol := tinypkg.NewPackage("m/component", "").NewSymbol("Provider")
			provider := &tinypkg.Var{Name: "provider", Symboler: providerSymbol}

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
