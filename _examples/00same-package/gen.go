// +build apikit

package main

import (
	"log"

	"m/00same-package/design"

	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/translate"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("!! %+v", err)
	}
}

func run() (err error) {
	emitter := emitgo.NewConfigFromRelativePath(design.ListUser, "..").NewEmitter()
	emitter.FilenamePrefix = "gen_" // generated file name is "gen_<name>.go"
	defer emitter.EmitWith(&err)

	config := translate.DefaultConfig()
	translator := translate.NewTranslator(config)

	rootpkg := emitter.RootPkg
	dst := rootpkg.Relative("runner", "")
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
	{
		here := dst
		code := translator.TranslateToInterface(here, "Component")
		emitter.Register(here, "component.go", code)
	}

	translator.Override("m", func() (*design.Messenger, error) { return nil, nil })
	return nil
}
