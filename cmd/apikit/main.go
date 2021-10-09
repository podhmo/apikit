package main

import (
	"log"
	"os"

	"github.com/podhmo/apikit/cmd/apikit/internal/clilib"
	"github.com/podhmo/apikit/cmd/apikit/internal/initcmd"
)

func main() {
	log.SetFlags(0)

	name := os.Args[0]
	subCommands := []*clilib.Command{
		initcmd.New(),
	}

	cmd := clilib.NewRouterCommand(name, subCommands)
	run := func() error {
		return cmd.Do([]*clilib.Command{cmd}, os.Args[1:])
	}
	if err := run(); err != nil {
		log.Fatalf("!! %+v", err)
	}
}
