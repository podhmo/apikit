package main

import (
	"os"

	"m/00simple/design"

	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/tinypkg"
	"github.com/podhmo/apikit/translate"
)

func main() {

	// TODO: use apikit
	resolver := resolve.NewResolver()
	translator := translate.NewTranslator(resolver, design.ListUser)

	here := tinypkg.NewPackage("m/00simple/component", "")
	code := translator.TranslateToInterface(here, "Component")

	code.Emit(os.Stdout, code)
}
