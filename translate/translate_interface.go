package translate

import (
	"fmt"
	"io"
	"log"
	"reflect"
	"strings"

	"github.com/podhmo/apikit/resolve"
)

type Need struct {
	Name string
	raw  resolve.Item
	rt   reflect.Type
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
					continue
				}
			}
			need := &Need{
				Name: arg.Name,
				rt:   k,
				raw:  arg,
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

func WriteInterface(w io.Writer, t *Tracker, name string) {
	fmt.Fprintf(w, "type %s interface {\n", name)
	usedNames := map[string]bool{}
	for _, need := range t.Needs {
		k := need.rt
		if len(t.seen[k]) > 1 {
			// TODO:
			// Db() *Db,  and Xdb() *Db
			panic("unsupported: TODO")
		} else {
			methodName := strings.ToUpper(string(need.Name[0])) + need.Name[1:] // TODO: use GoName
			typeExpr := need.raw                                                // xxx
			// TODO: T, (T, error)
			// TODO: support correct type expression
			fmt.Fprintf(w, "\t%s()%s\n", methodName, typeExpr)
			if _, duplicated := usedNames[methodName]; duplicated {
				log.Printf("WARN: method name %s is duplicated", methodName)
			}
			usedNames[methodName] = true
		}
	}
	io.WriteString(w, "}\n")
}
