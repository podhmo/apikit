package translate

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/podhmo/apikit/pkg/difftest"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
)

type Session struct{}
type Todo struct {
	Title string
	Done  bool
}

func AddTodo(session *Session, title string, done bool) (*Todo, error) {
	return nil, nil
}
func AddAlsoAnotherTodo(session *Session, anotherSession *Session, title string, done bool) (*Todo, error) {
	return nil, nil
}
func AddTodoWithContext(ctx context.Context, session *Session, title string, done bool) (*Todo, error) {
	return nil, nil
}
func MustAddTodo(session *Session, title string, done bool) *Todo {
	return nil
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
			name:  "RunAddAlsoAnotherTodo",
			input: AddAlsoAnotherTodo,
			here:  main,
			want: `
import (
	"github.com/podhmo/apikit/translate"
	"m/component"
)
func RunAddAlsoAnotherTodo(provider component.Provider, title string, done bool) (*translate.Todo, error) {
	var session *translate.Session
	{
		session = provider.Session()
	}
	var anotherSession *translate.Session
	{
		anotherSession = provider.AnotherSession()
	}
	return translate.AddAlsoAnotherTodo(session, anotherSession, title, done)
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
			name:  "RunAddTodoWithOverride1", // func()<T, error>
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
func RunAddTodoWithOverride1(provider component.Provider, title string, done bool) (*translate.Todo, error) {
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
		{
			name:  "RunAddTodoWithOverride2", // func()(<T>, func())
			input: AddTodo,
			here:  main,
			modifyTracker: func(tracker *Tracker) {
				rt := reflect.TypeOf(AddTodo).In(0)
				def := resolver.Def(func() (*Session, func()) { return nil, nil })
				tracker.Override(rt, "session", def)
			},
			want: `
import (
	"github.com/podhmo/apikit/translate"
	"m/component"
)
func RunAddTodoWithOverride2(provider component.Provider, title string, done bool) (*translate.Todo, error) {
	var session *translate.Session
	{
		var cleanup func()
		session, cleanup = provider.Session()
		if cleanup != nil {
			defer cleanup()
		}
	}
	return translate.AddTodo(session, title, done)
}`,
		},
		{
			name:  "RunAddTodoWithOverride3", // func()(<T>, func(), error)
			input: AddTodo,
			here:  main,
			modifyTracker: func(tracker *Tracker) {
				rt := reflect.TypeOf(AddTodo).In(0)
				def := resolver.Def(func() (*Session, func(), error) { return nil, nil, nil })
				tracker.Override(rt, "session", def)
			},
			want: `
import (
	"github.com/podhmo/apikit/translate"
	"m/component"
)
func RunAddTodoWithOverride3(provider component.Provider, title string, done bool) (*translate.Todo, error) {
	var session *translate.Session
	{
		var cleanup func()
		var err error
		session, cleanup, err = provider.Session()
		if cleanup != nil {
			defer cleanup()
		}
		if err != nil {
			return nil, err
		}
	}
	return translate.AddTodo(session, title, done)
}`,
		},
		{
			name:  "RunMustAddTodoWithOverride3", // func()(<T>, func(), error)
			input: MustAddTodo,
			here:  main,
			modifyTracker: func(tracker *Tracker) {
				rt := reflect.TypeOf(AddTodo).In(0)
				def := resolver.Def(func() (*Session, func(), error) { return nil, nil, nil })
				tracker.Override(rt, "session", def)
			},
			want: `
import (
	"github.com/podhmo/apikit/translate"
	"m/component"
)
func RunMustAddTodoWithOverride3(provider component.Provider, title string, done bool) *translate.Todo {
	var session *translate.Session
	{
		var cleanup func()
		var err error
		session, cleanup, err = provider.Session()
		if cleanup != nil {
			defer cleanup()
		}
		if err != nil {
			return nil
		}
	}
	return translate.MustAddTodo(session, title, done)
}`,
		},
		{
			name:  "RunAddTodoWithOverride4", // func(context.Context)(<T>, func(), error)
			input: AddTodo,
			here:  main,
			modifyTracker: func(tracker *Tracker) {
				rt := reflect.TypeOf(AddTodo).In(0)
				def := resolver.Def(func(ctx context.Context) (*Session, func(), error) { return nil, nil, nil })
				tracker.Override(rt, "session", def)
			},
			want: `
import (
	"context"
	"github.com/podhmo/apikit/translate"
	"m/component"
)
func RunAddTodoWithOverride4(ctx context.Context, provider component.Provider, title string, done bool) (*translate.Todo, error) {
	var session *translate.Session
	{
		var cleanup func()
		var err error
		session, cleanup, err = provider.Session(ctx)
		if cleanup != nil {
			defer cleanup()
		}
		if err != nil {
			return nil, err
		}
	}
	return translate.AddTodo(session, title, done)
}`,
		},
		// TODO: support consume function that returning zero value

		// // TODO: validation
		// // NG...
		// {
		// 	name:  "ngProviderPosision",
		// 	input: MustAddTodo,
		// 	here:  main,
		// 	modifyTracker: func(tracker *Tracker) {
		// 		rt := reflect.TypeOf(AddTodo).In(0)
		// 		def := resolver.Def(func() (*Session, error, func()) { return nil, nil, nil }) // not func()(<T>, func(), error)
		// 		tracker.Override(rt, "session", def)
		// 	},
		// },
		// {
		// 	name:  "ngProviderType",
		// 	input: MustAddTodo,
		// 	here:  main,
		// 	modifyTracker: func(tracker *Tracker) {
		// 		rt := reflect.TypeOf(AddTodo).In(0)
		// 		def := resolver.Def(func() (int, func(), error) { return 0, nil, nil }) // not *Session
		// 		tracker.Override(rt, "session", def)
		// 	},
		// },
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			translator := NewTranslator(resolver)
			if c.modifyTracker != nil {
				c.modifyTracker(translator.Tracker)
			}

			providerSymbol := tinypkg.NewPackage("m/component", "").NewSymbol("Provider")
			provider := &tinypkg.Var{Name: "provider", Node: providerSymbol}

			code := translator.TranslateToRunner(c.here, c.input, c.name, provider)
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
