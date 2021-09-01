package resolve

import (
	"github.com/podhmo/apikit/tinypkg"
	"github.com/podhmo/reflect-openapi/pkg/shape"
)

func ExtractSymbol(here *tinypkg.Package, s shape.Shape) *tinypkg.ImportedSymbol {
	pkg := tinypkg.NewPackage(s.GetPackage(), "")
	sym := pkg.NewSymbol(s.GetName())
	return here.Import(pkg).Lookup(sym)
}
