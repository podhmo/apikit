package main

import (
	"fmt"
	"log"

	"m/00same-package/design"

	"github.com/podhmo/apikit/pkg/emitfile"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/translate"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("!! %+v", err)
	}
}

func run() (retErr error) {
	rootdir := emitfile.DefinedDir(main)
	emitter := emitfile.New(rootdir)
	defer func() {
		if err := emitter.Emit(); err != nil {
			retErr = err
		}
	}()

	resolver := resolve.NewResolver()
	translator := translate.NewTranslator(resolver)
	dst := tinypkg.NewPackage("m/00same-package/runner", "")
	{
		pkg := dst
		code := translator.TranslateToRunner(pkg, design.ListUser, "", nil)
		emitter.Register(fmt.Sprintf("/runner/%s.go", code.Name), code)
	}
	{
		pkg := dst
		code := translator.TranslateToRunner(pkg, design.SendMessage, "", nil)
		emitter.Register(fmt.Sprintf("/runner/%s.go", code.Name), code)
	}
	{
		here := dst
		code := translator.TranslateToInterface(here, "Component")
		emitter.Register("/runner/component.go", code)
	}

	translator.Override("m", func() (*design.Messenger, error) { return nil, nil })
	return nil
}
