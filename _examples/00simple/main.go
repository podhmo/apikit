package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"m/00simple/design"

	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/tinypkg"
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

	here := tinypkg.NewPackage("m/00simple/component", "")

	{
		code := translator.TranslateToInterface(here, "Component")
		var buf bytes.Buffer
		if err := code.Emit(&buf, code); err != nil {
			return nil
		}
		writeFile("./00simple/component/component.go", buf.Bytes())
	}

	// TODO: detect provider name after emit code
	{
		pkg := tinypkg.NewPackage("m/00simple/runner", "")
		def := resolver.Def(design.ListUser)
		code := translator.TranslateToRunner(pkg, def, "", nil)
		var buf bytes.Buffer
		if err := code.Emit(&buf, code); err != nil {
			return nil
		}
		writeFile(fmt.Sprintf("./00simple/runner/%s.go", def.Name), buf.Bytes())
	}
	return nil
}

var mkdirSentinelMap = map[string]bool{}

func writeFile(path string, b []byte) error {
	if err := ioutil.WriteFile(path, b, 0666); err != nil {
		dirpath := filepath.Dir(path)
		if _, ok := mkdirSentinelMap[dirpath]; ok {
			return err
		}

		mkdirSentinelMap[dirpath] = true
		log.Printf("INFO: directory is not found, try to create %s", dirpath)
		if err := os.MkdirAll(dirpath, 0744); err != nil {
			log.Printf("ERROR: %s", err)
			return err
		}
		return ioutil.WriteFile(path, b, 0666)
	}
	return nil
}
