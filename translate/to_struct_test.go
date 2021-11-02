package translate

import (
	"fmt"
	"strings"
	"testing"
)

func TestStructFromFunc(t *testing.T) {
	config := DefaultConfig()
	config.Header = ""
	translator := NewTranslator(config)
	resolver := translator.Resolver

	fn := func(name string, age int) error { return nil }
	main := resolver.NewPackage("main", "")

	var buf strings.Builder
	code := translator.TranslateToStruct(main, fn, "S")
	if err := code.Emit(&buf); err != nil {
		t.Errorf("unexpected error, write %+v", err)
	}
	fmt.Println(buf.String())
}
