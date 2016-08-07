package thyme

import "strings"

type Tracker interface {
	Snap() (*Snapshot, error)
}

type Stream struct {
	Snapshots []*Snapshot
}

type Snapshot struct {
	Windows []*Window
	Active  int64
	Visible []int64
}

type Window struct {
	ID   int64
	Name string
}

type Winfo struct {
	App    string
	SubApp string
	Title  string
}

func (w *Window) Info() *Winfo {
	fields := strings.Split(w.Name, " - ")
	first := strings.TrimSpace(fields[0])
	last := strings.TrimSpace(fields[len(fields)-1])
	if last == "Google Chrome" {
		return &Winfo{
			App:    "Google Chrome",
			SubApp: strings.TrimSpace(fields[len(fields)-2]),
			Title:  strings.Join(fields[0:len(fields)-2], " - "),
		}
	} else if first == "Slack" {
		return &Winfo{
			App:    "Slack",
			SubApp: strings.TrimSpace(strings.Join(fields[1:], " - ")),
		}
	}
	return &Winfo{
		Title: w.Name,
	}
}
