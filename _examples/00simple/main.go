package main

import (
	"bytes"
	"fmt"
	"log"

	"m/00simple/design"
	"m/fileutil"

	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/tinypkg"
	"github.com/podhmo/apikit/translate"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("!! %+v", err)
	}
}

func run() error {
	resolver := resolve.NewResolver()

	translator := translate.NewTranslator(resolver, design.ListUser)
	translator.Override("m", func() (*design.Messenger, error) { return nil, nil })

	here := tinypkg.NewPackage("m/00simple/component", "")

	{
		code := translator.TranslateToInterface(here, "Component")
		var buf bytes.Buffer
		if err := code.Emit(&buf, code); err != nil {
			return nil
		}
		fileutil.WriteOrCreateFile("./00simple/component/component.go", buf.Bytes())
	}

	// TODO: detect provider name after emit code
	{
		pkg := tinypkg.NewPackage("m/00simple/runner", "")
		def := resolver.Def(design.ListUser)
		code := translator.TranslateToRunner(pkg, def, "", nil)
		var buf bytes.Buffer
		if err := code.Emit(&buf, code); err != nil {
			return nil
		}
		fileutil.WriteOrCreateFile(fmt.Sprintf("./00simple/runner/%s.go", def.Name), buf.Bytes())
	}
	{
		pkg := tinypkg.NewPackage("m/00simple/runner", "")
		def := resolver.Def(design.SendMessage)
		code := translator.TranslateToRunner(pkg, def, "", nil)
		var buf bytes.Buffer
		if err := code.Emit(&buf, code); err != nil {
			return nil
		}
		fileutil.WriteOrCreateFile(fmt.Sprintf("./00simple/runner/%s.go", def.Name), buf.Bytes())
	}
	return nil
}
