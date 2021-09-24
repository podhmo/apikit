package web

import (
	"fmt"
	"strings"

	"github.com/podhmo/apikit/resolve"
	reflectshape "github.com/podhmo/reflect-shape"
)

type PathVar struct {
	Name  string
	Shape reflectshape.Shape
	Regex string
}

type PathInfo struct {
	Name      string
	Shape     reflectshape.Function
	Variables []PathVar
}

var (
	ErrUnexpectedType            = fmt.Errorf("unexpected-type")
	ErrMismatchNumberOfVariables = fmt.Errorf("mismatch-number-of-variables")
)

func ExtractPathInfo(variableNames []string, def *resolve.Def) (*PathInfo, error) {
	vars := make([]PathVar, 0, len(variableNames))
	idx := 0
	for _, item := range def.Args {
		if item.Kind != resolve.KindPrimitive {
			continue
		}

		var regex string
		argname := item.Name
		if strings.Contains(argname, ":") {
			nameAndRegex := strings.SplitN(argname, ":", 2)
			argname = nameAndRegex[0]
			regex = nameAndRegex[1]
		}

		if len(variableNames) <= idx {
			got := make([]string, 0, len(vars))
			for _, v := range vars {
				got = append(got, v.Name)
			}
			return nil, fmt.Errorf("expected variables are %v, but want variables are %v (in def %s): %w", got, variableNames, def, ErrMismatchNumberOfVariables)
		}
		if argname != variableNames[idx] {
			continue
		}
		vars = append(vars, PathVar{Name: argname, Regex: regex, Shape: item.Shape})
		idx++
	}

	if len(vars) != len(variableNames) {
		got := make([]string, 0, len(vars))
		for _, v := range vars {
			got = append(got, v.Name)
		}
		return nil, fmt.Errorf("expected variables are %v, but want variables are %v (in def %s): %w", got, variableNames, def, ErrMismatchNumberOfVariables)
	}
	return &PathInfo{
		Name:      def.Shape.Name,
		Shape:     def.Shape,
		Variables: vars,
	}, nil
}
