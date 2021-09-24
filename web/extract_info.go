package web

import (
	"fmt"
	"strings"

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

func ExtractPathInfo(variableNames []string, shape reflectshape.Shape) (*PathInfo, error) {
	sfn, ok := shape.(reflectshape.Function)
	if !ok {
		return nil, fmt.Errorf("extracted shape of %q is not function : %w", shape, ErrUnexpectedType)
	}

	if sfn.Params.Len() < len(variableNames) {
		varCandidates := make([]string, 0, sfn.Params.Len())
		for _, argname := range sfn.Params.Keys {
			varCandidates = append(varCandidates, argname)
		}
		return nil, fmt.Errorf("extracted shape of %q 's variable candidates are %v, but want variables are %v : %w", shape, varCandidates, variableNames, ErrMismatchNumberOfVariables)
	}

	vars := make([]PathVar, 0, len(variableNames))
	idx := 0
	for i, argname := range sfn.Params.Keys {
		var regex string
		if strings.Contains(argname, ":") {
			nameAndRegex := strings.SplitN(argname, ":", 2)
			argname = nameAndRegex[0]
			regex = nameAndRegex[1]
		}
		if argname != variableNames[idx] {
			continue
		}
		vars = append(vars, PathVar{Name: argname, Regex: regex, Shape: sfn.Params.Values[i]})
		idx++
	}

	if len(vars) != len(variableNames) {
		got := make([]string, 0, len(vars))
		for _, v := range vars {
			got = append(got, v.Name)
		}
		return nil, fmt.Errorf("extracted shape of %q 's variables are %v, but want variables are %v : %w", shape, got, variableNames, ErrMismatchNumberOfVariables)
	}
	return &PathInfo{
		Name:      sfn.Name,
		Shape:     sfn,
		Variables: vars,
	}, nil
}
