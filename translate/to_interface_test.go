package translate

import (
	"fmt"
	"io"
	"strings"
	"testing"
)

type something struct{}
type CodeType string

func (s *something) Name() string                       { return "hello" }
func (s *something) Emit(w io.Writer) (CodeType, error) { return CodeType("misc"), nil }

func TestStructToInterface(t *testing.T) {
	ob := &something{}

	config := DefaultConfig()
	config.Header = ""
	translator := NewTranslator(config)
	resolver := translator.Resolver

	main := resolver.NewPackage("main", "")
	code := translator.TranslateToInterface(main, ob, "EmitTarget")

	var buf strings.Builder
	if err := code.Emit(&buf); err != nil {
		t.Fatalf("unexpected error %+v", err)
	}
	fmt.Println(buf.String())
}
