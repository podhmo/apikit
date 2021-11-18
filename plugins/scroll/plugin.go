package scroll

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/podhmo/apikit/code"
	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/plugins"
	"github.com/podhmo/apikit/resolve"
)

type Options struct {
	LatestIDTypeZeroValue interface{}
}

func (o Options) IncludeMe(pc *plugins.PluginContext, here *tinypkg.Package) error {
	return IncludeMe(
		pc.Config, pc.Resolver, pc.Emitter,
		here,
		o.LatestIDTypeZeroValue,
	)
}

func IncludeMe(
	config *code.Config, resolver *resolve.Resolver, emitter *emitgo.Emitter,
	here *tinypkg.Package,
	latestIDValue interface{},
) error {
	scrollT := resolver.Symbol(here, resolver.Shape(latestIDValue))
	if config.Verbose {
		config.Log.Printf("\t+ generate runtime-scroll [scrollT=%s]", scrollT)
	}

	// scroll.go
	c := &code.CodeEmitter{Code: config.NewCode(here, "scroll", func(w io.Writer, c *code.Code) error {
		c.AddDependency(scrollT)

		fpath := filepath.Join(emitgo.DefinedDir(IncludeMe), "internal/scroll.go")
		f, err := os.Open(fpath)
		if err != nil {
			return err
		}

		defer f.Close()
		r := bufio.NewReader(f)
		for {
			line, _, err := r.ReadLine()
			if err != nil {
				return err
			}
			if strings.HasPrefix(string(line), "package ") {
				break
			}
		}
		if _, err := io.Copy(w, r); err != nil {
			return err
		}

		fmt.Fprintln(w, "// todo: generics?")
		fmt.Fprintf(w, "type ScrollT = %s\n\n", scrollT)
		fmt.Fprintln(w, "")
		fmt.Fprintf(w, "func coerceScrollT(v reflect.Value) %s {\n", scrollT)
		switch reflect.TypeOf(latestIDValue).Kind() {
		case reflect.Int, reflect.Int32, reflect.Int16, reflect.Int64, reflect.Int8:
			fmt.Fprintf(w, "return %s(v.Int())\n", scrollT)
		case reflect.Uint, reflect.Uint32, reflect.Uint16, reflect.Uint64, reflect.Uint8:
			fmt.Fprintf(w, "return %s(v.Uint())\n", scrollT)
		case reflect.String:
			fmt.Fprintf(w, "return %s(v.String())\n", scrollT)
		default:
			rt := reflect.TypeOf(latestIDValue)
			return fmt.Errorf("unexpected latestValueType %v, kind=%v", rt, rt.Kind())
		}
		fmt.Fprintln(w, "}")
		return nil
	})}
	emitter.Register(here, c.Name, c)
	return nil
}
