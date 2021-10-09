package clilib

import (
	"flag"
	"fmt"
	"os"
)

type DoFunc func(path []*Command, args []string) error
type Command struct {
	*flag.FlagSet
	Do      DoFunc
	Options interface{}
}

func NewRouterCommand(name string, subcmds []*Command) *Command {
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage of %s:\n\n", name)
		fs.PrintDefaults()
		fmt.Fprintln(fs.Output(), "Available commands")
		for _, subcmd := range subcmds {
			fmt.Fprintf(fs.Output(), "  %s\n", subcmd.Name())
		}
	}
	return &Command{
		FlagSet: fs,
		Do: func(path []*Command, args []string) error {
			cmd := path[len(path)-1]
			if err := cmd.Parse(args); err != nil {
				return err
			}
			if cmd.NArg() == 0 {
				cmd.Usage()
				os.Exit(1)
				return nil
			}

			{
				args := cmd.Args()
				for _, subcmd := range subcmds {
					if subcmd.Name() == args[0] {
						return subcmd.Do(append(path, subcmd), args[1:])
					}
				}

				cmd.Usage()
				fmt.Fprintf(cmd.Output(), "\nunexpected command: %s\n", args[0])
				os.Exit(1)
				return nil
			}
		},
	}
}
