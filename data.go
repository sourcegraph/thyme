package thyme

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

type Tracker interface {
	Snap() (*Snapshot, error)
}

type Stream struct {
	Snapshots []*Snapshot
}

func (s Stream) Print() string {
	var b bytes.Buffer
	for _, snap := range s.Snapshots {
		fmt.Fprintf(&b, "%s", snap.Print())
	}
	return string(b.Bytes())
}

type Snapshot struct {
	Time    time.Time
	Windows []*Window
	Active  int64
	Visible []int64
}

func (s Snapshot) Print() string {
	var b bytes.Buffer

	var active *Window
	visible := make([]*Window, 0, len(s.Windows))
	other := make([]*Window, 0, len(s.Windows))
s_Windows:
	for _, w := range s.Windows {
		if w.ID == s.Active {
			active = w
			continue s_Windows
		}
		for _, v := range s.Visible {
			if w.ID == v {
				visible = append(visible, w)
				continue s_Windows
			}
		}
		other = append(other, w)
	}

	fmt.Fprintf(&b, "%s\n", s.Time.Format("Mon Jan 2 15:04:05 -0700 MST 2006"))
	if active != nil {
		fmt.Fprintf(&b, "\tActive: %s\n", active.Info().Print())
	}
	if len(visible) > 0 {
		fmt.Fprintf(&b, "\tVisible: ")
		for _, v := range visible {
			fmt.Fprintf(&b, "%s, ", v.Info().Print())
		}
		fmt.Fprintf(&b, "\n")
	}
	if len(other) > 0 {
		fmt.Fprintf(&b, "\tOther: ")
		for _, w := range other {
			fmt.Fprintf(&b, "%s, ", w.Info().Print())
		}
		fmt.Fprintf(&b, "\n")
	}
	return string(b.Bytes())
}

type Window struct {
	ID   int64
	Name string
}

var systemNames = map[string]struct{}{
	"XdndCollectionWindowImp": {},
	"unity-launcher":          {},
	"unity-panel":             {},
	"unity-dash":              {},
	"Hud":                     {},
	"Desktop":                 {},
}

// IsSystem returns true if the window is a system window (like
// "unity-panel" and thus shouldn't be considered an application
// visible to the end-users)
func (w *Window) IsSystem() bool {
	if _, is := systemNames[w.Name]; is {
		return true
	}
	return false
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

type Winfo struct {
	App    string
	SubApp string
	Title  string
}

func (w Winfo) Print() string {
	return fmt.Sprintf("[%s|%s|%s]", w.App, w.SubApp, w.Title)
}
