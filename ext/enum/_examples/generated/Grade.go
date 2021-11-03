// Code generated by "github.com/podhmo/apikit"; DO NOT EDIT.

package generated

import (
	"fmt"
)

type Grade string

const (
	GradeS Grade = "s"
	GradeA Grade = "a"
	GradeB Grade = "b"
	GradeC Grade = "c"
	GradeD Grade = "d"
)

var ErrNoGrade = fmt.Errorf("no Grade")

func (v Grade) Validate() error {
	switch v {
	case GradeS, GradeA, GradeB, GradeC, GradeD:
		return nil
	default:
		return fmt.Errorf("unexpected value %v: %w", ErrNoGrade)
	}
}

func MustGrade(v string) Grade {
	retval := Grade(v)
	if err := retval.Validate(); err != nil {
		panic(err)
	}
	return retval
}
