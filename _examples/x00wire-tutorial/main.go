package main

import (
	"fmt"
	"log"

	"github.com/podhmo/apikit/pkg/emitfile"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/translate"

	"m/x00wire-tutorial/action"
	"m/x00wire-tutorial/tutorial"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("!! %+v", err)
	}
}

func run() (retErr error) {
	rootdir := "./x00wire-tutorial"
	emitter := emitfile.New(rootdir)
	defer func() {
		if err := emitter.Emit(); err != nil {
			retErr = err
		}
	}()

	resolver := resolve.NewResolver()
	translator := translate.NewTranslator(resolver)
	dst := tinypkg.NewPackage("m/x00wire-tutorial/runner", "")

	{
		here := tinypkg.NewPackage("m/x00wire-tutorial/component", "")
		code := translator.TranslateToInterface(here, "Component")
		emitter.Register("/component/component.go", code)
	}
	{
		pkg := dst
		code := translator.TranslateToRunner(pkg, action.StartEvent, "", nil)
		emitter.Register(fmt.Sprintf("/runner/%s.go", code.Name), code)
	}

	translator.Override("ev", tutorial.NewEvent)
	return nil
}
