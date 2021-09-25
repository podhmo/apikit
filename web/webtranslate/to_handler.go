package webtranslate

import (
	"io"

	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/web"
)

func WriteHandlerFunc(w io.Writer, here *tinypkg.Package, resolver *resolve.Resolver, info *web.PathInfo) {
	http := resolver.NewPackage("net/http", "")

	args := []*tinypkg.Var{
		{Name: "w", Node: http.NewSymbol("ResponseWriter")},
		{Name: "req", Node: &tinypkg.Pointer{Lv: 1, V: http.NewSymbol("Request")}},
	}
	f := here.NewFunc(info.Def.Name, args, nil)
	tinypkg.WriteFunc(w, here, "", f, func() error {
		return nil
	})
}
