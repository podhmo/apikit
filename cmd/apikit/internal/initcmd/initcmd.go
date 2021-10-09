package initcmd

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"

	"github.com/podhmo/apikit/cmd/apikit/internal/clilib"
	"github.com/podhmo/apikit/code"
	"github.com/podhmo/apikit/pkg/emitgo"
)

func New() *clilib.Command {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	var options struct {
		Silent bool `json:"silent"`
	}
	fs.BoolVar(&options.Silent, "silent", false, "silent option")

	fs.Usage = func() {
		name := fs.Name()
		fmt.Fprintf(fs.Output(), "Usage of %s <rootpkg>:\n\n", name)
		fs.PrintDefaults()
		fmt.Fprintln(fs.Output(), "  rootpkg")
		fmt.Fprintln(fs.Output(), "\tthe package path of root package")
	}

	return &clilib.Command{
		FlagSet: fs,
		Options: &options,
		Do: func(path []*clilib.Command, args []string) (err error) {
			cmd := path[len(path)-1]
			if err := cmd.Parse(args); err != nil {
				return err
			}
			if cmd.NArg() < 1 {
				cmd.Usage()
				return fmt.Errorf("init <rootpkg>")
			}

			rootpkgPath := cmd.Args()[0]
			cfg := code.DefaultConfig()
			cfg.Verbose = !options.Silent
			cfg.Header = ""

			rootpkg := cfg.Resolver.NewPackage(rootpkgPath, "main")
			designpkg := rootpkg.Relative("design", "")
			actionpkg := rootpkg.Relative("action", "")

			emitter := emitgo.New(filepath.Base(rootpkgPath), rootpkg)
			emitter.FileEmitter.Config.Verbose = cfg.Verbose
			defer emitter.EmitWith(&err)

			{
				here := designpkg
				c := cfg.NewCode(here, "code.go", func(w io.Writer, c *code.Code) error {
					c.Import(cfg.Resolver.NewPackage("github.com/morikuni/failure", ""))
					source := `
// error codes for your application.
const (
	NotFound        failure.StringCode = "NotFound"
	Forbidden       failure.StringCode = "Forbidden"
	ValidationError failure.StringCode = "ValidationError"
)

func HTTPStatusOf(err error) int {
	if err == nil {
		return 200 // http.StatusOK
	}

	c, ok := failure.CodeOf(err)
	if !ok {
		return 500 // http.StatusInternalServerError
	}
	switch c {
	case NotFound:
		return 404 // http.StatusNotFound
	case Forbidden:
		return 403 // http.StatusForbidden
	case ValidationError:
		return 422 // http.StatusUnprocessableEntity // or http.StatusBadRequest?
	default:
		return 500 // http.StatusInternalServerError
	}
}`
					fmt.Fprintln(w, source)
					return nil
				})
				emitter.Register(here, c.Name, &code.CodeEmitter{Code: c})
			}
			{
				here := actionpkg
				c := cfg.NewCode(here, "Hello.go", func(w io.Writer, c *code.Code) error {
					c.Import(cfg.Resolver.NewPackage("context", ""))
					c.Import(cfg.Resolver.NewPackage("log", ""))
					c.Import(cfg.Resolver.NewPackage("os", ""))
					source := `
type HelloOutput struct {
	Message string` + "`json: \"message\"`" + `
}
func Hello(ctx context.Context, logger *log.Logger) (*HelloOutput, error) {
	logger.Printf("hello")
	return &HelloOutput{Message: "hello"}, nil
}

func NewLogger() (*log.Logger, error) {
	return log.New(os.Stderr, "app ", 0), nil
}
`
					fmt.Fprintln(w, source)
					return nil
				})
				emitter.Register(here, c.Name, &code.CodeEmitter{Code: c})
			}
			{
				here := rootpkg
				c := cfg.NewCode(here, "gen.go", func(w io.Writer, c *code.Code) error {
					c.Import(cfg.Resolver.NewPackage("log", ""))
					c.Import(cfg.Resolver.NewPackage("context", ""))

					c.Import(cfg.Resolver.NewPackage("github.com/podhmo/apikit/pkg/emitgo", ""))
					c.Import(cfg.Resolver.NewPackage("github.com/podhmo/apikit/web", ""))
					c.Import(cfg.Resolver.NewPackage("github.com/podhmo/apikit/web/webgen/gen-chi", "genchi"))

					c.Import(designpkg)
					c.Import(actionpkg)
					source := `
// generate code: VERBOSE=1 go run gen.go

func main() {
	if err := run(); err != nil {
		log.Fatalf("!! %+v", err)
	}
}

func run() (err error) {
	emitter := emitgo.NewFromRelativePath(action.Hello, "..")
	defer emitter.EmitWith(&err)

	r := web.NewRouter()
	r.Get("/hello", action.Hello)

	c := genchi.DefaultConfig()
	// c.Override("logger", action.NewLogger)

	g := c.New(emitter)
	return g.Generate(
		context.Background(),
		r,
		design.HTTPStatusOf,
	)
}
`
					fmt.Fprintln(w, source)
					return nil
				})
				c.Header = `// +build apikit

`
				emitter.Register(here, c.Name, &code.CodeEmitter{Code: c})
			}
			return nil
		},
	}
}
