package webtranslate

import (
	"fmt"
	"io"

	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/web"
)

func WriteHandlerFunc(w io.Writer, here *tinypkg.Package, name string, resolver *resolve.Resolver, info *web.PathInfo, runtime *tinypkg.Package, getProviderFunc *tinypkg.Var) error {
	args := []*tinypkg.Var{
		getProviderFunc,
	}
	if name == "" {
		name = info.Def.Name
	}
	f := here.NewFunc(name, args, nil)
	return tinypkg.WriteFunc(w, here, "", f, func() error {
		// TODO: import http
		fmt.Fprintf(w, "\treturn func(w http.ResponseWriter, req *http.Request) http.HandlerFunc{\n")
		fmt.Fprintf(w, "\t\tresult, err := %s()\n",
			tinypkg.ToRelativeTypeString(here, info.Def.Symbol),
		)
		fmt.Fprintf(w, "\t\treturn %s(w, req, result, err)\n", tinypkg.ToRelativeTypeString(here, runtime.NewSymbol("HandleResult")))
		defer fmt.Fprintln(w, "\t}")
		return nil
	})
}

// type GetProviderFunc func(*http.Request) (*http.Request, Provider, error)

// func ListMessageHandler(getProvider GetProviderFunc) http.HandlerFunc {
// 	return func(w http.ResponseWriter, req *http.Request) {
// 		req, provider, err := getProvider(req)
// 		if err != nil {
// 			webruntime.HandleResult(w, req, nil, err)
// 			return
// 		}

// 		var db *DB
// 		{
// 			db = provider.DB()
// 		}

// 		result, err := ListMesesage(db)
// 		webruntime.HandleResult(w, req, result, err)
// 	}
// }
