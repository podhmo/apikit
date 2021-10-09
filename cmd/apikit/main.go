package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/podhmo/apikit/cmd/apikit/internal/clilib"
)

func main() {
	log.SetFlags(0)

	name := os.Args[0]
	subCommands := []*clilib.Command{
		NewInitCommand(),
	}

	cmd := clilib.NewRouterCommand(name, subCommands)
	run := func() error {
		return cmd.Do([]*clilib.Command{cmd}, os.Args[1:])
	}
	if err := run(); err != nil {
		log.Fatalf("!! %+v", err)
	}
}

func NewInitCommand() *clilib.Command {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	var options struct {
		Verbose bool `json:"verbose"`
	}
	fs.BoolVar(&options.Verbose, "verbose", false, "verbose option")

	fs.Usage = func() {
		name := fs.Name()
		fmt.Fprintf(fs.Output(), "Usage of %s <path>:\n\n", name)
		fs.PrintDefaults()
		fmt.Fprintln(fs.Output(), "  path")
		fmt.Fprintln(fs.Output(), "\tthe path of directory")
	}

	return &clilib.Command{
		FlagSet: fs,
		Options: &options,
		Do: func(path []*clilib.Command, args []string) error {
			cmd := path[len(path)-1]
			if err := cmd.Parse(args); err != nil {
				return err
			}
			if cmd.NArg() < 1 {
				cmd.Usage()
				return fmt.Errorf("init <path>")
			}

			fmt.Println("@@", "hello", options.Verbose)
			return nil
		},
	}
}
