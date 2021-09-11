package tinypkg

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/podhmo/apikit/difftest"
)

func TestWalk(t *testing.T) {
	u := NewUniverse()
	pkg := u.NewPackage("m/foo", "foo")
	main := u.NewPackage("main", "main")

	cases := []struct {
		msg   string
		here  *Package
		input Symboler
		want  string
	}{
		{
			msg:   "same pkg, simple symbol",
			here:  pkg,
			input: pkg.NewSymbol("Foo"),
			want: `
Foo
`,
		},
		{
			msg:   "another pkg, simple symbol",
			here:  main,
			input: pkg.NewSymbol("Foo"),
			want: `
foo.Foo
`,
		},
		{
			msg:   "map[Foo]Bar",
			here:  main,
			input: &Map{K: pkg.NewSymbol("Foo"), V: pkg.NewSymbol("Bar")},
			want: `
foo.Foo
foo.Bar
`,
		},
		{
			msg:  "map[*Foo][]map[Bar][3]Boo",
			here: main,
			input: &Map{
				K: &Pointer{1, pkg.NewSymbol("Foo")},
				V: &Slice{V: &Map{K: pkg.NewSymbol("Bar"), V: &Array{N: 3, V: pkg.NewSymbol("Boo")}}}},
			want: `
foo.Foo
foo.Bar
foo.Boo
`,
		},
		{
			msg:  "func(ctx context.Context, userID string)([]*User, error)",
			here: main,
			input: &Func{
				Params: []*Var{
					{Name: "ctx", Symboler: u.NewPackage("context", "").NewSymbol("Context")},
					{Name: "userId", Symboler: builtins.NewSymbol("string")},
				},
				Returns: []*Var{
					{Symboler: &Slice{V: &Pointer{Lv: 1, V: pkg.NewSymbol("User")}}},
					{Symboler: builtins.NewSymbol("error")},
				},
			},
			want: `
context.Context
string
foo.User
error
`,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.msg, func(t *testing.T) {
			buf := new(bytes.Buffer)
			seen := map[string]bool{}

			if err := Walk(c.input, func(sym *Symbol) error {
				importedSymbol := c.here.Import(sym.Package).Lookup(sym)
				s := importedSymbol.String()
				if _, existed := seen[s]; existed {
					return nil
				}
				seen[s] = true
				fmt.Fprintln(buf, s)
				return nil
			}); err != nil {
				t.Fatalf("unexpected error %+v", err)
			}

			if got, want := strings.TrimSpace(buf.String()), strings.TrimSpace(c.want); got != want {
				difftest.LogDiffGotStringAndWantString(t, got, want)
			}
		})
	}
}
