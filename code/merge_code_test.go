package code

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/podhmo/apikit/pkg/difftest"
	"github.com/podhmo/apikit/pkg/tinypkg"
)

func TestMergeCode(t *testing.T) {
	config := DefaultConfig()
	config.Header = ""
	pkg := tinypkg.NewPackage("foo", "")

	foo := config.NewCode(pkg, "Foo", func(w io.Writer, c *Code) error {
		c.Import(tinypkg.NewPackage("fmt", "fmt"))
		fmt.Fprintln(w, `
func Foo() {
	fmt.Println("foo")
}
		`)
		return nil
	})
	bar := config.NewCode(pkg, "Foo", func(w io.Writer, c *Code) error {
		c.Import(tinypkg.NewPackage("fmt", "fmt"))
		c.Import(tinypkg.NewPackage("strings", "strings"))
		fmt.Fprintln(w, `
func Bar() {
	Foo()
	fmt.Println(strings.Repeat("bar", 2))
}
		`)
		return nil
	})

	t.Run("base", func(t *testing.T) {
		var buf strings.Builder
		if err := (&CodeEmitter{foo}).Emit(&buf); err != nil {
			t.Errorf("unexpected error: %+v", err)
			return
		}
		want := `
package foo

import (
	"fmt"
)


func Foo() {
	fmt.Println("foo")
}`
		if want, got := difftest.NormalizeString(want), difftest.NormalizeString(buf.String()); want != got {
			difftest.LogDiffGotStringAndWantString(t, want, got)
		}
	})

	t.Run("merged", func(t *testing.T) {
		foo.AddDependency(bar) // side-effect!

		var buf strings.Builder
		if err := (&CodeEmitter{foo}).Emit(&buf); err != nil {
			t.Errorf("unexpected error: %+v", err)
			return
		}
		want := `
package foo

import (
	"fmt"
	"strings"
)


func Foo() {
	fmt.Println("foo")
}


func Bar() {
	Foo()
	fmt.Println(strings.Repeat("bar", 2))
}
`
		if want, got := difftest.NormalizeString(want), difftest.NormalizeString(buf.String()); want != got {
			difftest.LogDiffGotStringAndWantString(t, want, got)
		}
	})
}

func TestMergeCodeWithCodeEmitter(t *testing.T) {
	config := DefaultConfig()
	config.Header = ""
	pkg := tinypkg.NewPackage("foo", "")

	code := config.ZeroCode(pkg, "code")
	code.AddDependency(config.NewCode(pkg, "foo", func(w io.Writer, c *Code) error {
		fmt.Fprintln(w, "// with Code")
		return nil
	}))
	code.AddDependency(&CodeEmitter{Code: config.NewCode(pkg, "bar", func(w io.Writer, c *Code) error {
		fmt.Fprintln(w, "// with CodeEmitter")
		return nil
	})},
	)

	var buf bytes.Buffer
	if err := (&CodeEmitter{Code: code}).Emit(&buf); err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	want := `
package foo


// with Code
// with CodeEmitter
`
	if want, got := strings.TrimSpace(want), strings.TrimSpace(buf.String()); want != got {
		difftest.LogDiffGotStringAndWantString(t, got, want)
	}
}
