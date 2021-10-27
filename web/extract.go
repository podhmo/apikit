package web

import (
	"fmt"
	"strings"

	"github.com/podhmo/apikit/pkg/namelib"
	"github.com/podhmo/apikit/resolve"
	reflectshape "github.com/podhmo/reflect-shape"
)

type PathVar struct {
	Name  string // not argname (e.g. /{articleId} and func(articleID string) { ... })
	Regex string
	Shape reflectshape.Shape
}

type PathInfo struct {
	Def      *resolve.Def
	VarNames []string
	Vars     map[string]*PathVar // argname -> PathVar
}

var (
	ErrUnexpectedType            = fmt.Errorf("unexpected-type")
	ErrMismatchNumberOfVariables = fmt.Errorf("mismatch-number-of-variables")
)

func ExtractPathInfo(pathParameters []string, def *resolve.Def) (*PathInfo, error) {
	argNames := make([]string, 0, len(pathParameters))
	vars := make(map[namelib.Normalized]*PathVar, len(pathParameters))
	for _, item := range def.Args {
		if item.Kind != resolve.KindPrimitive {
			continue
		}

		argname := item.Name
		normalized := namelib.ToNormalized(argname)

		argNames = append(argNames, argname)
		vars[normalized] = &PathVar{
			Name:  argname,
			Shape: item.Shape,
		}
	}

	if len(argNames) != len(pathParameters) {
		want := argNames
		got := pathParameters
		return nil, fmt.Errorf("path-parameters are %v, but extracted values are %v (in %s): %w", got, want, def, ErrMismatchNumberOfVariables)
	}

	missingParams := make([]string, 0, len(pathParameters))
	for _, name := range pathParameters {
		var regex string
		if strings.Contains(name, ":") {
			nameAndRegex := strings.SplitN(name, ":", 2)
			name = nameAndRegex[0]
			regex = nameAndRegex[1]
		}
		normalized := namelib.ToNormalized(name)
		item, ok := vars[normalized]
		if !ok {
			missingParams = append(missingParams, name)
			continue
		}

		item.Name = name
		item.Regex = regex
	}

	if len(missingParams) > 0 {
		want := argNames
		got := missingParams
		return nil, fmt.Errorf("parameters %v are missing (arguments=%v, in %s): %w", got, want, def, ErrMismatchNumberOfVariables)
	}

	argVars := make(map[string]*PathVar)
	for _, name := range argNames {
		argVars[name] = vars[namelib.ToNormalized(name)]
	}
	return &PathInfo{
		Def:      def,
		VarNames: argNames,
		Vars:     argVars,
	}, nil
}
