package main

import (
	"log"

	"m/02value-to-interface/db"

	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/translate"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("!! %+v", err)
	}
}

func run() (err error) {
	emitter := emitgo.NewConfigFromRelativePath(db.NewDB, "..").NewEmitter()
	defer emitter.EmitWith(&err)

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
}
