package resolve

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/podhmo/apikit/pkg/tinypkg"
)

type Need struct {
	rt          reflect.Type
	OverrideDef *Def

	*Item
}

type Tracker struct {
	Needs []*Need

	visitedDef map[string]bool
	Seen       map[reflect.Type][]*Need
}

func NewTracker() *Tracker {
	return &Tracker{
		visitedDef: map[string]bool{},
		Seen:       map[reflect.Type][]*Need{},
	}
}

func (t *Tracker) Track(def *Def) {
	path := def.Shape.GetFullName()
	if _, ok := t.visitedDef[path]; ok {
		return
	}
	t.visitedDef[path] = true
toplevel:
	for _, arg := range def.Args {
		arg := arg
		switch arg.Kind {
		case KindIgnored, KindUnsupported:
			continue
		case KindData:
			continue
		case KindComponent:
			k := arg.Shape.GetReflectType()
			needs := t.Seen[k]

			for _, n := range needs {
				if n.Name == arg.Name {
					continue toplevel
				}
			}
			need := &Need{
				rt:   k,
				Item: &arg,
			}
			t.Seen[k] = append(t.Seen[k], need)
			t.Needs = append(t.Needs, need)
		case KindPrimitive:
			continue
		default:
			panic(fmt.Sprintf("unexpected kind %s", arg.Kind))
		}
	}
}

func (t *Tracker) Override(rt reflect.Type, name string, def *Def) (prev *Def) {
	for {
		if rt.Kind() != reflect.Ptr {
			break
		}
		rt = rt.Elem()
	}

	k := rt
	var target *Need
	for _, need := range t.Seen[k] {
		if need.Name == name {
			target = need
			break
		}
	}
	if target == nil {
		if name == "" && len(t.Seen[k]) == 1 {
			target = t.Seen[k][0]
		} else {
			target = &Need{
				rt: k,
				Item: &Item{
					Kind:  KindComponent,
					Name:  name,
					Shape: def.Shape.Returns.Values[0], // xxx:
				},
			}
			t.Seen[k] = append(t.Seen[k], target)
			t.Needs = append(t.Needs, target)
		}
	}
	prev = target.OverrideDef
	target.OverrideDef = def
	return prev
}

func (t *Tracker) ExtractInterface(here *tinypkg.Package, resolver *Resolver, name string) *tinypkg.Interface {
	usedNames := map[string]bool{}
	methods := make([]*tinypkg.Func, 0, len(t.Needs))
	for _, need := range t.Needs {
		k := need.rt

		methodName := need.rt.Name()
		if len(t.Seen[k]) > 1 {
			methodName = strings.ToUpper(string(need.Name[0])) + need.Name[1:] // TODO: use GoName
		}
		shape := need.Shape
		if need.OverrideDef != nil {
			shape = need.OverrideDef.Shape
		}

		sym := resolver.Symbol(here, shape)
		m, ok := sym.(*tinypkg.Func)
		if !ok {
			m = here.NewFunc(methodName, nil, []*tinypkg.Var{{Node: sym}})
		}
		m.Name = methodName

		if _, duplicated := usedNames[methodName]; duplicated {
			log.Printf("WARN: method name %s is duplicated", methodName)
		}
		usedNames[methodName] = true
		methods = append(methods, m)
	}
	return here.NewInterface(name, methods)
}
