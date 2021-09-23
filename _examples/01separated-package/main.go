package main

import (
	"log"

	"m/01separated-package/design"

	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/translate"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("!! %+v", err)
	}
}

func run() (err error) {
	emitter := emitgo.NewFromRelativePath(design.ListUser, "..")
	defer emitter.EmitWith(&err)

	config := translate.DefaultConfig()
	translator := translate.NewTranslator(config)

	rootpkg := emitter.RootPkg
	dst := rootpkg.Relative("runner", "")

	{
		here := rootpkg.Relative("component", "")
		code := translator.TranslateToInterface(here, "Component")
		emitter.Register(here, "component.go", code)
	}
	{
		pkg := dst
		code := translator.TranslateToRunner(pkg, design.ListUser, "", nil)
		emitter.Register(pkg, code.Name, code)
	}
	{
		pkg := dst
		code := translator.TranslateToRunner(pkg, design.SendMessage, "", nil)
		emitter.Register(pkg, code.Name, code)
	}

	translator.Override("", func() (*design.Messenger, error) { return nil, nil })
	return nil
}
