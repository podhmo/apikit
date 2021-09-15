package main

import (
	"bytes"
	"fmt"
	"log"

	"m/00same-package/design"
	"m/fileutil"

	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
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
	dst := tinypkg.NewPackage("m/00same-package/runner", "")

	{
		here := dst
		code := translator.TranslateToInterface(here, "Component")
		var buf bytes.Buffer
		if err := code.Emit(&buf, code); err != nil {
			return nil
		}
		fileutil.WriteOrCreateFile("./00same-package/runner/component.go", buf.Bytes())
	}

	// TODO: detect provider name after emit code
	{
		pkg := dst
		def := resolver.Def(design.ListUser)
		code := translator.TranslateToRunner(pkg, def, "", nil)
		var buf bytes.Buffer
		if err := code.Emit(&buf, code); err != nil {
			return nil
		}
		fileutil.WriteOrCreateFile(fmt.Sprintf("./00same-package/runner/%s.go", def.Name), buf.Bytes())
	}
	{
		pkg := dst
		def := resolver.Def(design.SendMessage)
		code := translator.TranslateToRunner(pkg, def, "", nil)
		var buf bytes.Buffer
		if err := code.Emit(&buf, code); err != nil {
			return nil
		}
		fileutil.WriteOrCreateFile(fmt.Sprintf("./00same-package/runner/%s.go", def.Name), buf.Bytes())
	}
	return nil
}
