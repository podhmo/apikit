package resolve

import (
	"fmt"
	"strings"
	"testing"

	reflectshape "github.com/podhmo/reflect-shape"
)

func TestStructFromFunc(t *testing.T) {
	resolver := NewResolver()
	fn := func(name string, age int) error { return nil }
	main := resolver.NewPackage("main", "")

	shape, err := StructFromShape(resolver, resolver.Shape(fn).(reflectshape.Function)) //
	if err != nil {
		t.Fatalf("unexpected error %+v", err)
	}

	s := toStruct(main, resolver, "S", shape.(reflectshape.Struct))
	var buf strings.Builder
	if err := WriteStruct(&buf, main, "", s); err != nil {
		t.Errorf("unexpected error, write %+v", err)
	}
	fmt.Println(buf.String())
}
