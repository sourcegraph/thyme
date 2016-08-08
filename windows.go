package thyme

import "fmt"

type WindowsTracker struct{}

var _ Tracker = (*WindowsTracker)(nil)

func (t *WindowsTracker) Deps() string {
	return "TODO"
}

func (t *WindowsTracker) Snap() (*Snapshot, error) {
	return nil, fmt.Errorf("TODO")
}
