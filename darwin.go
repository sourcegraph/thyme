package thyme

import "fmt"

type DarwinTracker struct{}

var _ Tracker = (*DarwinTracker)(nil)

func (t *DarwinTracker) Deps() string {
	return "TODO"
}

func (t *DarwinTracker) Snap() (*Snapshot, error) {
	return nil, fmt.Errorf("TODO")
}
