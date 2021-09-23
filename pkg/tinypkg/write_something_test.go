package tinypkg

import (
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
