package main

import (
	"m/02value-to-interface/db"

	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/translate"
)

func main() {
	emitgo.NewConfigFromRelativePath(db.NewDB, "..").MustEmitWith(func(emitter *emitgo.Emitter) error {
		config := translate.DefaultConfig()
		translator := translate.NewTranslator(config)

		rootpkg := emitter.RootPkg
		dst := rootpkg.Relative("component", "")
		{
			pkg := dst
			code := translator.TranslateToInterface(pkg, &db.DB{}, "DB")
			emitter.Register(pkg, code.Name, code)
		}
		return nil
	})
}
