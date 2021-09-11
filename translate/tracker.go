package translate

import (
	"fmt"
	"reflect"

	"github.com/podhmo/apikit/resolve"
)

type Need struct {
	rt reflect.Type
	resolve.Item
}

type Tracker struct {
	Needs []*Need

	visitedDef map[string]bool
	seen       map[reflect.Type][]*Need
}

func NewTracker() *Tracker {
	return &Tracker{
		visitedDef: map[string]bool{},
		seen:       map[reflect.Type][]*Need{},
	}
}

func (t *Tracker) Track(def *resolve.Def) {
	path := def.Shape.GetFullName()
	if _, ok := t.visitedDef[path]; ok {
		return
	}
	t.visitedDef[path] = true
toplevel:
	for _, arg := range def.Args {
		arg := arg
		switch arg.Kind {
		case resolve.KindIgnored, resolve.KindUnsupported:
			continue
		case resolve.KindData:
			continue
		case resolve.KindComponent:
			k := arg.Shape.GetReflectType()
			needs := t.seen[k]

			for _, n := range needs {
				if n.Name == arg.Name {
					continue toplevel
				}
			}
			need := &Need{
				rt:   k,
				Item: arg,
			}
			t.seen[k] = append(t.seen[k], need)
			t.Needs = append(t.Needs, need)
		case resolve.KindPrimitive:
			continue
		default:
			panic(fmt.Sprintf("unexpected kind %s", arg.Kind))
		}
	}
}
