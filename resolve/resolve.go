package resolve

import (
	"reflect"

	"github.com/podhmo/apikit/tinypkg"
	reflectshape "github.com/podhmo/reflect-shape"
	"github.com/podhmo/reflect-shape/arglist"
)

type Resolver struct {
	extractor *reflectshape.Extractor
	universe  *tinypkg.Universe
}

func NewResolver() *Resolver {
	e := &reflectshape.Extractor{
		Seen:           map[reflect.Type]reflectshape.Shape{},
		ArglistLookup:  arglist.NewLookup(),
		RevisitArglist: true,
	}
	return &Resolver{
		extractor:    e,
		universe:     tinypkg.NewUniverse(),
		symbolsCache: map[reflectshape.Shape][]*symbolCacheItem{},
	}
}

func (r *Resolver) Resolve(fn interface{}) *Def {
	return ExtractDef(r.universe, r.extractor, fn)
}
