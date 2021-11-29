package code

import (
	"bytes"
	"fmt"
	"io"

	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
)

const (
	PriorityFirst   = -10
	PrioritySecond  = -1
	PriorityDefault = 0
)

type EmitCodeFunc func(w io.Writer, e *CodeEmitter) error
type Config struct {
	*resolve.Config

	Header        string
	DisableFormat bool

	EmitCodeFunc EmitCodeFunc
	Resolver     *resolve.Resolver
}

func DefaultConfig() *Config {
	resolver := resolve.NewResolver()
	c := &Config{
		Header:        Header,
		DisableFormat: false,
		Resolver:      resolver,
		Config:        resolver.Config,
	}
	c.EmitCodeFunc = c.defaultEmitCodeFunc
	return c
}

const Header = `// Code generated by "github.com/podhmo/apikit"; DO NOT EDIT.

`

func (c *Config) defaultEmitCodeFunc(w io.Writer, code *CodeEmitter) error {
	if code.Header != "" {
		fmt.Fprintln(w, code.Header)
	} else {
		fmt.Fprintln(w, c.Header)
	}

	fmt.Fprintf(w, "package %s\n\n", code.Here.Name)

	// first, emit code
	buf := new(bytes.Buffer)
	seen := make(map[codeLike]bool, len(code.Code.depends))
	if err := c.defaultEmitCodeInner(buf, code.Code, seen); err != nil {
		return fmt.Errorf("emit code in code %q : %w", code.Name, err)
	}

	// second, emit imports (for dependencies)
	collector := tinypkg.NewImportCollector(code.Here)
	if err := code.CollectImports(collector); err != nil {
		return fmt.Errorf("collect import in code %q : %w ", code.Name, err)
	}
	imports := collector.Imports
	if err := code.EmitImports(w, imports); err != nil {
		if err != ErrNoImports {
			return fmt.Errorf("emit import in code %q : %w ", code.Name, err)
		}
	} else {
		io.WriteString(w, "\n")
	}

	if _, err := io.Copy(w, buf); err != nil {
		return fmt.Errorf("copy body in code %q, something wrong: %w", code.Name, err)
	}
	return nil
}

type codeLike interface {
	Dependencies() []tinypkg.Node
	EmitContent(w io.Writer) error
}

func (c *Config) defaultEmitCodeInner(w io.Writer, code codeLike, seen map[codeLike]bool) error {
	if _, ok := seen[code]; ok {
		return nil
	}
	seen[code] = true
	if err := code.EmitContent(w); err != nil {
		return err
	}
	for _, x := range code.Dependencies() {
		if code, ok := x.(codeLike); ok {
			if err := c.defaultEmitCodeInner(w, code, seen); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Config) NewCode(
	here *tinypkg.Package,
	name string,
	emitCode func(io.Writer, *Code) error,
) *Code {
	return &Code{
		Name:     name,
		Here:     here,
		EmitCode: emitCode,
		Config:   c,
	}
}
