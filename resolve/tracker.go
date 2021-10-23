package resolve

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/podhmo/apikit/pkg/tinypkg"
	reflectshape "github.com/podhmo/reflect-shape"
)

type Need struct {
	rt          reflect.Type
	OverrideDef *Def

	*Item
}

type Tracker struct {
	Resolver *Resolver
	Needs    []*Need

	visitedDef map[string]bool
	seen       map[reflect.Type][]*Need
}

func NewTracker(resolver *Resolver) *Tracker {
	return &Tracker{
		Resolver:   resolver,
		visitedDef: map[string]bool{},
		seen:       map[reflect.Type][]*Need{},
	}
}

func (t *Tracker) Track(def *Def) {
	path := def.Shape.GetFullName()
	if _, ok := t.visitedDef[path]; ok {
		return
	}
	t.visitedDef[path] = true
	for _, arg := range def.Args {
		arg := arg
		switch arg.Kind {
		case KindIgnored, KindUnsupported:
			continue
		case KindData:
			continue
		case KindComponent:
			t.track(arg)
		case KindPrimitive, KindPrimitivePointer:
			continue
		default:
			panic(fmt.Sprintf("unexpected kind %s", arg.Kind))
		}
	}
}

func (t *Tracker) track(arg Item) {
	k := arg.Shape.GetReflectType()
	needs := t.seen[k]

	for _, n := range needs {
		if n.Name == arg.Name {
			return
		}
	}
	need := &Need{
		rt:   k,
		Item: &arg,
	}
	t.seen[k] = append(t.seen[k], need)
	t.Needs = append(t.Needs, need)
}

func (t *Tracker) Override(name string, providerFunc interface{}) (prev *Def, err error) {
	rt := reflect.TypeOf(providerFunc)
	if rt.Kind() != reflect.Func {
		return nil, fmt.Errorf("unexpected providerFunc, only function %v", rt)
	}

	targetType := rt.Out(0)
	def := t.Resolver.Def(providerFunc)
	if t.Resolver.Config.Verbose {
		t.Resolver.Config.Log.Printf("\t! override provider[name=%q] = %s.%s [type=%s]", name, def.Package.Path, def.Symbol, rt)
	}
	return t.overrideByDef(targetType, name, def), nil
}

func (t *Tracker) overrideByDef(rt reflect.Type, name string, def *Def) (prev *Def) {
	for {
		if rt.Kind() != reflect.Ptr {
			break
		}
		rt = rt.Elem()
	}

	k := rt
	var target *Need
	for _, need := range t.seen[k] {
		if need.Name == name {
			target = need
			break
		}
	}
	if target == nil {
		if name == "" && len(t.seen[k]) == 1 {
			target = t.seen[k][0]
		} else {
			target = &Need{
				rt: k,
				Item: &Item{
					Kind:  KindComponent,
					Name:  name,
					Shape: def.Shape.Returns.Values[0], // xxx:
				},
			}
			t.seen[k] = append(t.seen[k], target)
			t.Needs = append(t.Needs, target)
		}
	}
	prev = target.OverrideDef
	target.OverrideDef = def

	if len(def.Args) > 0 {
		for _, x := range def.Args {
			if x.Kind == KindComponent {
				t.track(x)
			}
		}
	}

	return prev
}

func (t *Tracker) ExtractInterface(here *tinypkg.Package, resolver *Resolver, name string) *tinypkg.Interface {
	usedNames := map[string]bool{}
	methods := make([]*tinypkg.Func, 0, len(t.Needs))
	for _, need := range t.Needs {
		methodName := t.ExtractMethodName(need.rt, need.Name)
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

func (t *Tracker) ExtractMethodName(rt reflect.Type, name string) string {
	methodName := rt.Name()
	if len(t.seen[rt]) > 1 {
		methodName = strings.ToUpper(string(name[0])) + name[1:] // TODO: use GoName
	}
	return methodName
}

func (t *Tracker) ExtractComponentFactoryShape(x Item) reflectshape.Shape {
	shape := x.Shape
	for _, need := range t.seen[x.Shape.GetReflectType()] {
		if need.Name == x.Name && need.OverrideDef != nil {
			shape = need.OverrideDef.Shape
			break
		}
	}
	return shape
}
