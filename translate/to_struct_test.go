package translate

import (
	"strings"
	"testing"

	"github.com/podhmo/apikit/pkg/difftest"
	"github.com/podhmo/apikit/pkg/tinypkg"
)

type Color string

func TestStructFromFunc(t *testing.T) {
	config := DefaultConfig()
	config.Header = ""
	translator := NewTranslator(config)
	resolver := translator.Resolver
	main := resolver.NewPackage("main", "")

	type Point struct {
		X int `json:"x"`
		Y int `json:"y"`
	}
	cases := []struct {
		msg   string
		here  *tinypkg.Package
		input interface{}
		want  string
	}{
		{
			msg:   "funcToStruct",
			here:  main,
			input: func(name string, age int) error { return nil },
			want: `package main


type S struct {
	Name string ` + "`json:\"name\"`" + `
	Age int ` + "`json:\"age\"`" + `
}
			`,
		},
		{
			msg:   "funcToStruct-with-new-type",
			here:  main,
			input: func(name string, age int, color Color) error { return nil },
			want: `package main

import (
	"github.com/podhmo/apikit/translate"
)

type S struct {
	Name string ` + "`json:\"name\"`" + `
	Age int ` + "`json:\"age\"`" + `
	Color translate.Color ` + "`json:\"color\"`" + `
}
		`,
		},
		{
			msg:   "funcToStruct-with-struct",
			here:  main,
			input: func(point Point, verbose *bool) error { return nil },
			want: `package main

import (
	"github.com/podhmo/apikit/translate"
)

type S struct {
	translate.Point
	Verbose *bool ` + "`json:\"verbose\"`" + `
}
		`,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.msg, func(t *testing.T) {
			var buf strings.Builder
			code := translator.TranslateToStruct(c.here, c.input, "S")

			if err := code.Emit(&buf); err != nil {
				t.Errorf("unexpected error, write %+v", err)
			}
			if want, got := strings.TrimSpace(c.want), strings.TrimSpace(buf.String()); want != got {
				difftest.LogDiffGotStringAndWantString(t, got, want)
			}
		})
	}
}
