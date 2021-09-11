package resolve

import (
	"reflect"
	"sync"

	"github.com/podhmo/apikit/tinypkg"
	reflectshape "github.com/podhmo/reflect-shape"
	"github.com/podhmo/reflect-shape/arglist"
)

type Resolver struct {
	extractor *reflectshape.Extractor
	universe  *tinypkg.Universe

	mu           sync.RWMutex
	symbolsCache map[reflectshape.Identity][]*symbolCacheItem
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
		symbolsCache: map[reflectshape.Identity][]*symbolCacheItem{},
	}
}

func (r *Resolver) Resolve(fn interface{}) *Def {
	return ExtractDef(r.universe, r.extractor, fn)
}

type symbolCacheItem struct {
	Here     *tinypkg.Package
	Shape    reflectshape.Shape
	Symboler tinypkg.Symboler
}

func (r *Resolver) Symbol(here *tinypkg.Package, s reflectshape.Shape) tinypkg.Symboler {
	r.mu.RLock()
	k := s.GetIdentity()
	cached, ok := r.symbolsCache[k]
	r.mu.RUnlock()

	if ok {
		for _, item := range cached {
			if item.Here == here {
				return item.Symboler
			}
		}
	}
	symboler := ExtractSymbol(here, s)

	r.mu.Lock()
	r.symbolsCache[k] = append(cached, &symbolCacheItem{
		Here:     here,
		Shape:    s,
		Symboler: symboler,
	})
	r.mu.Unlock()

	return symboler
}
