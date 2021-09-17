package main

import (
	"fmt"
	"log"

	"m/01separated-package/design"

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
	rootdir := "./01separated-package"
	emitter := emitfile.New(rootdir)
	defer func() {
		if err := emitter.Emit(); err != nil {
			retErr = err
		}
	}()

	resolver := resolve.NewResolver()
	translator := translate.NewTranslator(resolver)
	dst := tinypkg.NewPackage("m/01separated-package/runner", "")

	{
		here := tinypkg.NewPackage("m/01separated-package/component", "")
		code := translator.TranslateToInterface(here, "Component")
		emitter.Register("/component/component.go", code)
	}
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

	translator.Override("m", func() (*design.Messenger, error) { return nil, nil })
	return nil
}