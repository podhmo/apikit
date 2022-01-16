package action

import (
	"m/x00wire-tutorial/tutorial"
)

func StartEvent(ev *tutorial.Event, phrase string) error {
	ev.Start()
	return nil
}
