// Code generated by "github.com/podhmo/apikit"; DO NOT EDIT.

package runner

import (
	"m/x00wire-tutorial/component"
	"m/x00wire-tutorial/tutorial"
)

func StartEvent(component component.Component, phrase string) error {
	var ev tutorial.Event
	{
		var err error
		ev, err = component.Event()
		if err != nil {
			return err
		}
	}
	return action.StartEvent(ev, phrase)
}
