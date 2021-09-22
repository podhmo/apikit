package apikit

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"unsafe"

	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/translate"
)

func GenerateTypeCode(here *tinypkg.Package, target interface{}) (tinypkg.Node, *translate.Code) {
	resolver := resolve.NewResolver() // xxx: omit

	shape := resolver.Shape(target)
	targetTypeSymbol := resolver.Symbol(here, shape)

	code := &translate.Code{
		Here: here,
		ImportPackages: func() ([]*tinypkg.ImportedPackage, error) {
			var imports []*tinypkg.ImportedPackage
			seen := map[*tinypkg.Package]bool{}
			use := func(sym *tinypkg.Symbol) error {
				if sym.Package.Path == "" {
					return nil // bultins type (e.g. string, bool, ...)
				}
				if _, ok := seen[sym.Package]; ok {
					return nil
				}
				seen[sym.Package] = true
				if here == sym.Package {
					return nil
				}
				imports = append(imports, here.Import(sym.Package))
				return nil
			}
			if err := tinypkg.Walk(targetTypeSymbol, use); err != nil {
				return nil, err
			}
			return imports, nil
		},
		EmitCode: func(w io.Writer) error {
			_, err := fmt.Fprintf(w, "type T = %s\n", targetTypeSymbol)
			return err
		},
	}
	return targetTypeSymbol, code // Todo: code has symbol
}

func GenerateRouterCode(pkg *tinypkg.Package) *translate.Code {
	filename := emitgo.DefinedFile(NewRouter)
	code := &translate.Code{
		Here: pkg,
		EmitCode: func(w io.Writer) error {
			f, err := os.Open(filename)
			if err != nil {
				return fmt.Errorf("generate router, read file, %w", err)
			}

			// skip package syntax
			r := bufio.NewReader(f)
			{
				for {
					line, _, err := r.ReadLine()
					if err != nil {
						return fmt.Errorf("generate router, readline, %w", err)
					}
					if strings.HasPrefix(*(*string)(unsafe.Pointer(&line)), "package ") {
						break
					}
				}
			}

			if _, err := io.Copy(w, r); err != nil {
				return fmt.Errorf("generate router, copy, %w", err)
			}
			defer f.Close()
			return nil
		},
	}
	return code
}
