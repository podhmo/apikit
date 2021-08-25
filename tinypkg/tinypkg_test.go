package tinypkg

import (
	"fmt"
	"testing"
)

func TestSymbol(t *testing.T) {
	u := NewUniverse()
	xxx := u.NewPackage("m/xxx", "xxx")
	Foo := xxx.NewSymbol("Foo")

	{
		main := u.NewPackage("main", "main")
		mainFoo := main.Import(xxx).Lookup(Foo)
		Bar := main.NewSymbol("Bar")

		t.Run("imported package", func(t *testing.T) {
			got := fmt.Sprintf("%s", mainFoo)
			want := "xxx.Foo"
			if want != got {
				t.Errorf("\nwant:\n\t%q\nbut got:\n\t%q", want, got)
			}
		})
		t.Run("qualified imported package", func(t *testing.T) {
			mainFoo := main.ImportAs(xxx, "x").Lookup(Foo)
			got := fmt.Sprintf("%s", mainFoo)
			want := "x.Foo"
			if want != got {
				t.Errorf("\nwant:\n\t%q\nbut got:\n\t%q", want, got)
			}
		})
		t.Run("same package", func(t *testing.T) {
			got := fmt.Sprintf("%s", Bar)
			want := "Bar"
			if want != got {
				t.Errorf("\nwant:\n\t%q\nbut got:\n\t%q", want, got)
			}
		})
	}
}
