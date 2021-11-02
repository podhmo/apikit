package translate

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/podhmo/apikit/pkg/difftest"
	"github.com/podhmo/apikit/pkg/tinypkg"
)

type something struct{}
type CodeType string

func (s *something) Name() string                       { return "hello" }
func (s *something) Emit(w io.Writer) (CodeType, error) { return CodeType("misc"), nil }

func TestStructToInterface(t *testing.T) {
	config := DefaultConfig()
	config.Header = ""
	translator := NewTranslator(config)
	resolver := translator.Resolver

	main := resolver.NewPackage("main", "")
	pkg := resolver.NewPackage(reflect.TypeOf(something{}).PkgPath(), "")

	cases := []struct {
		msg   string
		here  *tinypkg.Package
		input interface{}
		want  string
	}{
		{
			msg:   "same-package",
			here:  pkg,
			input: &something{},
			want: `package translate

import (
	"io"
)

type Generated interface {
	Emit(io.Writer) (CodeType, error)
	Name() string
}`,
		},
		{
			msg:   "other-package",
			here:  main,
			input: &something{},
			want: `package main

import (
	"io"
	"github.com/podhmo/apikit/translate"
)

type Generated interface {
	Emit(io.Writer) (translate.CodeType, error)
	Name() string
}`,
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.msg, func(t *testing.T) {
			code := translator.TranslateToInterface(c.here, c.input, "Generated")
			var buf strings.Builder
			if err := code.Emit(&buf); err != nil {
				t.Fatalf("unexpected error %+v", err)
			}
			if want, got := strings.TrimSpace(c.want), strings.TrimSpace(buf.String()); want != got {
				difftest.LogDiffGotStringAndWantString(t, got, want)
			}
		})
	}
}
