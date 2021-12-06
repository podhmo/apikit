package graph

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/podhmo/apikit/pkg/namelib"
)

type Config struct {
	RootQueryName    string
	RootMutationName string
	NameTag          string
}

type Router struct {
	names   []string
	Config  *Config
	Schemas map[string]*Schema
	mu      sync.Mutex
}

func NewRouter() *Router {
	return &Router{
		Config: &Config{
			RootQueryName:    "query",
			RootMutationName: "mutation",
			NameTag:          "json",
		},
		Schemas: map[string]*Schema{},
	}
}

func (r *Router) Query(fields ...*SchemaField) *Schema {
	s, _ := r.getOrCreateSchema("Query", "Query", fields)
	return s
}

func (r *Router) Mutation(fields ...*SchemaField) *Schema {
	s, _ := r.getOrCreateSchema("Mutation", "Mutation", fields)
	return s
}

func (r *Router) getOrCreateSchema(id string, name string, fields []*SchemaField) (*Schema, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	created := false
	s, ok := r.Schemas[id]
	if !ok {
		s = &Schema{ID: id, Name: name}
		created = true
		r.Schemas[name] = s
		r.names = append(r.names, id)
	}

	// TODO: dedup
	// log.Printf("WARNING: in schema=%q, field %q is already existed", s.Name, f.Name)
	s.Fields = append(s.Fields, fields...)
	return s, created
}

func (r *Router) Object(ob interface{}, options ...*SchemaField) (*Schema, error) {
	rt := reflect.TypeOf(ob)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	if rt.Kind() != reflect.Struct {
		return nil, fmt.Errorf("unexpected type %v, please give me an object type value", rt)
	}

	k := rt.String()
	s, created := r.getOrCreateSchema(k, rt.Name(), options)
	if created {
		fields := make([]*SchemaField, rt.NumField())
		nametag := r.Config.NameTag
		for i := 0; i < rt.NumField(); i++ {
			rf := rt.Field(i)
			fieldname := rf.Tag.Get(nametag)
			if fieldname == "" {
				fieldname = namelib.ToUnexported(rf.Name)
			}
			fields[i] = &SchemaField{
				GoName:     rf.Name,
				Name:       fieldname,
				Type:       rf.Type,
				HasResolve: false,
			}
		}
		r.mu.Lock()
		s.Fields = append(fields, s.Fields...)
		r.mu.Unlock()
	}
	return s, nil
}

func (r *Router) MustObject(ob interface{}, fields ...*SchemaField) *Schema {
	s, err := r.Object(ob, fields...)
	if err != nil {
		panic(err)
	}
	return s
}

func (r *Router) Field(name string, resolver interface{}) *SchemaField {
	rfn := reflect.TypeOf(resolver)
	return &SchemaField{
		GoName:     namelib.ToExported(name),
		Name:       name,
		Type:       rfn.Out(0),
		Resolve:    resolver,
		HasResolve: true,
	}
}

type Schema struct {
	ID      string
	Name    string
	Fields  []*SchemaField
	Ignored []string
}

type SchemaField struct {
	GoName     string
	Name       string
	Type       reflect.Type // xxx
	Resolve    interface{}  `json:"-"` // func
	HasResolve bool
}

// TODO: implements
// TODO: Ignore
// TODO: Enum
// TODO: Union

//
// type Router interface {
// 	Query(options ...Option)
// 	Mutation(options ...Option)

// 	Object(ob interface{}, options ...Option)
// 	Implements(ob interface{}) Option
// 	Field(name string, r Resolver) Option
// 	Ignore(names ...string) Option

// 	// TODO
// 	Union(fn interface{})
// 	Enum(values ...interface{})
// }
