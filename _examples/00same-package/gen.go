//go:build apikit
// +build apikit

package main

import (
	"m/00same-package/design"

	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/translate"
)

func main() {
	emitgo.NewConfigFromRelativePath(design.ListUser, "..").MustEmitWith(func(emitter *emitgo.Emitter) error {
		emitter.FilenamePrefix = "gen_" // generated file name is "gen_<name>.go"

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
			code := translator.ExtractProviderInterface(here, "Component")
			emitter.Register(here, "component.go", code)
		}

		translator.Override("m", func() (*design.Messenger, error) { return nil, nil })
		return nil
	})
}
