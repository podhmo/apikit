// Code generated by "github.com/podhmo/apikit"; DO NOT EDIT.

package component

import (
	"m/x00wire-tutorial/tutorial"
)

type Component interface {
	Event(g tutorial.Greeter) (tutorial.Event, error)
}
