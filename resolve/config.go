package resolve

import (
	"log"
	"os"
	"strconv"
)

type Logger interface {
	Printf(fmt string, args ...interface{})
}

type Config struct {
	Log     Logger
	Verbose bool
	Debug   bool

	IgnoreMap map[string]bool
}

func DefaultConfig() *Config {
	verbose := false
	if v, err := strconv.ParseBool(os.Getenv("VERBOSE")); err == nil {
		verbose = v
	}
	debug := false
	if v, err := strconv.ParseBool(os.Getenv("DEBUG")); err == nil {
		debug = v
	}

	return &Config{
		Log:       log.New(os.Stderr, "", 0),
		Verbose:   verbose,
		Debug:     debug,
		IgnoreMap: map[string]bool{"context.Context": true},
	}
}
