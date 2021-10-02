package resolve

import (
	"testing"

	"github.com/podhmo/apikit/pkg/tinypkg"
)

type Request struct{}

type ProviderT interface{}

type GetProvider struct {
	T ProviderT

	GetProvider func(*Request) (*Request, ProviderT, error)
	unexported  func()
}

func TestNewPreModule(t *testing.T) {
	resolver := NewResolver()
	target := GetProvider{}

	pm, err := resolver.PreModule(target)
	if err != nil {
		t.Fatalf("new pre module %+v", err)
	}

	t.Run("name", func(t *testing.T) {
		if want, got := "GetProvider", pm.Name; want != got {
			t.Errorf("want name:\n\t%q\nbut got:\n\t%q", want, got)
		}
	})
	t.Run("args", func(t *testing.T) {
		if want, got := 1, len(pm.Args); want != got {
			t.Errorf("want len(args):\n\t%d\nbut got:\n\t%d", want, got)
		}
	})
	t.Run("funcs", func(t *testing.T) {
		if want, got := 2, len(pm.Funcs); want != got {
			t.Errorf("want len(args):\n\t%d\nbut got:\n\t%d", want, got)
		}
	})
}

func TestNewModule(t *testing.T) {
	here := tinypkg.NewPackage("main", "")

	resolver := NewResolver()
	target := GetProvider{}
	pm, err := resolver.PreModule(target)
	if err != nil {
		t.Fatalf("new pre module %+v", err)
	}

	m, err := pm.NewModule(here, here.NewSymbol("Provider"))
	if err != nil {
		t.Fatalf("new module %+v", err)
	}

	t.Run("string", func(t *testing.T) {
		want := `Module[here=main, args=[Provider], funcs=[GetProvider unexported]]`
		got := m.String()
		if want != got {
			t.Errorf("want String()\n\t%s\nbut got\n\t%s", want, got)
		}
	})
	t.Run("funcs", func(t *testing.T) {
		name := "GetProvider"
		want := `func(*resolve.Request) (*resolve.Request, Provider, error)`
		f, err := m.Func(name)
		if err != nil {
			t.Errorf("Func() %+v", err)
		}
		if got := f.String(); want != got {
			t.Errorf("want String()\n\t%s\nbut got\n\t%s", want, got)
		}
	})
}
