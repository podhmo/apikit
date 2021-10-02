package tinypkg

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/podhmo/apikit/pkg/difftest"
)

func TestWriteFunc(t *testing.T) {
	pkg := NewPackage("m/foo", "")
	main := NewPackage("main", "")

	fn := pkg.NewFunc(
		"DoSomething",
		[]*Var{{Name: "ctx", Node: NewPackage("context", "").NewSymbol("Context")}},
		[]*Var{
			{Node: &Pointer{Lv: 1, V: pkg.NewSymbol("Foo")}},
			{Node: NewSymbol("error")},
		},
	)

	cases := []struct {
		msg   string
		input *Func
		here  *Package
		want  string
	}{
		{
			msg:   "same package",
			input: fn,
			here:  pkg,
			want: `
func DoSomething(ctx context.Context) (*Foo, error) {
	return nil, nil
}`,
		},
		{
			msg:   "another package",
			input: fn,
			here:  main,
			want: `
func DoSomething(ctx context.Context) (*foo.Foo, error) {
	return nil, nil
}`,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.msg, func(t *testing.T) {
			var buf strings.Builder
			if err := WriteFunc(&buf, c.here, "", fn, func() error {
				fmt.Fprintln(&buf, "\treturn nil, nil")
				return nil
			}); err != nil {
				t.Fatalf("unexpected error %+v", err)
			}
			if want, got := strings.TrimSpace(c.want), strings.TrimSpace(buf.String()); want != got {
				difftest.LogDiffGotStringAndWantString(t, got, want)
			}
		})
	}
}

func TestWriteInterface(t *testing.T) {
	pkg := NewPackage("net/http", "")
	main := NewPackage("main", "")

	iface := pkg.NewInterface(
		"Handler",
		[]*Func{
			pkg.NewFunc("ServeHTTP", []*Var{
				{Node: pkg.NewSymbol("ResponseWriter")},
				{Node: &Pointer{Lv: 1, V: pkg.NewSymbol("Request")}},
			}, nil),
		},
	)

	cases := []struct {
		msg   string
		input *Interface
		here  *Package
		want  string
	}{
		{
			msg:   "same package",
			input: iface,
			here:  pkg,
			want: `
type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}`,
		},
		{
			msg:   "another package",
			input: iface,
			here:  main,
			want: `
type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}`,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.msg, func(t *testing.T) {
			var buf strings.Builder
			if err := WriteInterface(&buf, c.here, "", iface); err != nil {
				t.Fatalf("unexpected error %+v", err)
			}
			if want, got := strings.TrimSpace(c.want), strings.TrimSpace(buf.String()); want != got {
				difftest.LogDiffGotStringAndWantString(t, got, want)
			}
		})
	}
}

func TestWriteBinding(t *testing.T) {
	main := NewPackage("main", "")

	// func(ctx context.Context) (*DB, func(), error) { ... }
	newDB := main.NewFunc(
		"NewDB",
		[]*Var{{Name: "ctx", Node: NewPackage("context", "").NewSymbol("Context")}},
		[]*Var{
			{Node: &Pointer{Lv: 1, V: main.NewSymbol("DB")}},
			{Node: &Func{}},
			{Node: NewSymbol("error")},
		},
	)

	cases := []struct {
		msg     string
		here    *Package
		binding *Binding
		returns []*Var

		want    string
		wantErr error
	}{
		{
			msg:  "ok-provider-return-1--external-return-1",
			here: main,
			binding: &Binding{
				Name: "db",
				// func(ctx context.Context) *DB { ... }
				Provider: &Func{Name: newDB.Name, Args: newDB.Args,
					Returns: []*Var{newDB.Returns[0]},
				},
			},
			returns: []*Var{
				{Node: &Pointer{Lv: 1, V: main.NewSymbol("Foo")}},
			},
			want: `
var db *DB
{
	db = NewDB(ctx)
}
`,
		},
		{
			msg:  "ok-provider-return-1--external-return-1--provider-alias",
			here: main,
			binding: &Binding{
				Name: "db",
				// func(ctx context.Context) *DB { ... }
				Provider: &Func{Name: newDB.Name, Args: newDB.Args,
					Returns: []*Var{newDB.Returns[0]},
				},
				ProviderAlias: "MustNewDB",
			},
			returns: []*Var{
				{Node: &Pointer{Lv: 1, V: main.NewSymbol("Foo")}},
			},
			want: `
var db *DB
{
	db = MustNewDB(ctx)
}
`,
		},
		{
			msg:  "ok-provider-return-2-with-error--external-return-1",
			here: main,
			binding: &Binding{
				Name:     "db",
				HasError: true,
				// func(ctx context.Context) (*DB, error) { ... }
				Provider: &Func{Name: newDB.Name, Args: newDB.Args,
					Returns: []*Var{newDB.Returns[0], newDB.Returns[2]},
				},
			},
			returns: []*Var{
				{Node: &Pointer{Lv: 1, V: main.NewSymbol("Foo")}},
			},
			want: `
var db *DB
{
	var err error
	db, err = NewDB(ctx)
	if err != nil {
		return nil
	}
}
`,
		},
		{
			msg:  "ok-provider-return-2-with-cleanup--external-return-1",
			here: main,
			binding: &Binding{
				Name:       "db",
				HasCleanup: true,
				// func(ctx context.Context) (*DB, func()) { ... }
				Provider: &Func{Name: newDB.Name, Args: newDB.Args,
					Returns: []*Var{newDB.Returns[0], newDB.Returns[1]},
				},
			},
			returns: []*Var{
				{Node: &Pointer{Lv: 1, V: main.NewSymbol("Foo")}},
			},
			want: `
var db *DB
{
	var cleanup func()
	db, cleanup = NewDB(ctx)
	if cleanup != nil {
		defer cleanup()
	}
}
`,
		},
		{
			msg:  "ok-provider-return-3--external-return-1",
			here: main,
			binding: &Binding{
				Name:       "db",
				HasError:   true,
				HasCleanup: true,
				// func(ctx context.Context) (*DB, func(), error) { ... }
				Provider: &Func{Name: newDB.Name, Args: newDB.Args,
					Returns: newDB.Returns,
				},
			},
			returns: []*Var{
				{Node: &Pointer{Lv: 1, V: main.NewSymbol("Foo")}},
			},
			want: `
var db *DB
{
	var cleanup func()
	var err error
	db, cleanup, err = NewDB(ctx)
	if cleanup != nil {
		defer cleanup()
	}
	if err != nil {
		return nil
	}
}
`,
		},
		{
			msg:  "ng-provider-return-0",
			here: main,
			binding: &Binding{
				Name: "db",
				// func(ctx context.Context) (*DB, func(), error) { ... }
				Provider: &Func{Name: newDB.Name, Args: newDB.Args,
					Returns: newDB.Returns,
				},
			},
			returns: nil,
			wantErr: ErrUnexpectedReturnType,
		},
		{
			msg:  "ng-provider-return-3",
			here: main,
			binding: &Binding{
				Name: "db",
				// func(ctx context.Context) (*DB, func(), error) { ... }
				Provider: &Func{Name: newDB.Name, Args: newDB.Args,
					Returns: newDB.Returns,
				},
			},
			returns: []*Var{
				{Node: &Pointer{Lv: 1, V: main.NewSymbol("Foo")}},
			},
			wantErr: ErrUnexpectedReturnType,
		},
		// the variation for number of external returns
		{
			msg:  "ok-provider-return-2-with-error--external-return-0",
			here: main,
			binding: &Binding{
				Name:     "db",
				HasError: true,
				// func(ctx context.Context) (*DB, error) { ... }
				Provider: &Func{Name: newDB.Name, Args: newDB.Args,
					Returns: []*Var{newDB.Returns[0], newDB.Returns[2]},
				},
				ZeroReturnsDefault: "panic(err) // TODO: fix-it",
			},
			returns: []*Var{},
			want: `
var db *DB
{
	var err error
	db, err = NewDB(ctx)
	if err != nil {
		panic(err) // TODO: fix-it
	}
}
`,
		},
		{
			msg:  "ok-provider-return-2-with-error--external-return-2-with-error",
			here: main,
			binding: &Binding{
				Name:     "db",
				HasError: true,
				// func(ctx context.Context) (*DB, error) { ... }
				Provider: &Func{Name: newDB.Name, Args: newDB.Args,
					Returns: []*Var{newDB.Returns[0], newDB.Returns[2]},
				},
			},
			returns: []*Var{
				{Node: &Pointer{Lv: 1, V: main.NewSymbol("Foo")}},
				{Node: NewSymbol("error")},
			},
			want: `
var db *DB
{
	var err error
	db, err = NewDB(ctx)
	if err != nil {
		return nil, err
	}
}
`,
		},
		{
			msg:  "ok-provider-return-2-with-error--external-return-2",
			here: main,
			binding: &Binding{
				Name:     "db",
				HasError: true,
				// func(ctx context.Context) (*DB, error) { ... }
				Provider: &Func{Name: newDB.Name, Args: newDB.Args,
					Returns: []*Var{newDB.Returns[0], newDB.Returns[2]},
				},
			},
			returns: []*Var{
				{Node: &Pointer{Lv: 1, V: main.NewSymbol("Foo")}},
				{Node: NewSymbol("func()")},
			},
			want: `
var db *DB
{
	var err error
	db, err = NewDB(ctx)
	if err != nil {
		return nil, nil
	}
}
`,
		},
		{
			msg:  "ok-provider-return-2-with-error--external-return-3-with-error",
			here: main,
			binding: &Binding{
				Name:     "db",
				HasError: true,
				// func(ctx context.Context) (*DB, error) { ... }
				Provider: &Func{Name: newDB.Name, Args: newDB.Args,
					Returns: []*Var{newDB.Returns[0], newDB.Returns[2]},
				},
			},
			returns: []*Var{
				{Node: &Pointer{Lv: 1, V: main.NewSymbol("Foo")}},
				{Node: NewSymbol("func()")},
				{Node: NewSymbol("error")},
			},
			want: `
var db *DB
{
	var err error
	db, err = NewDB(ctx)
	if err != nil {
		return nil, nil, err
	}
}
`,
		},
		{
			msg:  "ok-provider-return-2-with-error--external-return-3",
			here: main,
			binding: &Binding{
				Name:     "db",
				HasError: true,
				// func(ctx context.Context) (*DB, error) { ... }
				Provider: &Func{Name: newDB.Name, Args: newDB.Args,
					Returns: []*Var{newDB.Returns[0], newDB.Returns[2]},
				},
			},
			returns: []*Var{
				{Node: &Pointer{Lv: 1, V: main.NewSymbol("Foo")}},
				{Node: NewSymbol("func()")},
				{Node: NewSymbol("func()")},
			},
			want: `
var db *DB
{
	var err error
	db, err = NewDB(ctx)
	if err != nil {
		return nil, nil, nil
	}
}
`,
		},
		{
			msg:  "ng-provider-return-2-with-error--external-return-4",
			here: main,
			binding: &Binding{
				Name:     "db",
				HasError: true,
				// func(ctx context.Context) (*DB, error) { ... }
				Provider: &Func{Name: newDB.Name, Args: newDB.Args,
					Returns: []*Var{newDB.Returns[0], newDB.Returns[2]},
				},
			},
			returns: []*Var{
				{Node: &Pointer{Lv: 1, V: main.NewSymbol("Foo")}},
				{Node: NewSymbol("func()")},
				{Node: NewSymbol("func()")},
				{Node: NewSymbol("func()")},
			},
			wantErr: ErrUnexpectedExternalReturnType,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.msg, func(t *testing.T) {
			buf := new(strings.Builder)
			err := c.binding.WriteWithCleanupAndError(buf, c.here, "", c.returns)
			if c.wantErr == nil {
				if err != nil {
					t.Fatalf("unexpected error: %+v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("must error is occured. (want error is %+v)", c.wantErr)
				}
				if !errors.Is(err, c.wantErr) {
					t.Fatalf("unexpected error %+v (want error is %+v)", err, c.wantErr)
				}
				return
			}

			if want, got := strings.TrimSpace(c.want), strings.TrimSpace(buf.String()); want != got {
				difftest.LogDiffGotStringAndWantString(t, got, want)
			}
		})
	}
}
