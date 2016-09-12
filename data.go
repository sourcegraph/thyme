package thyme

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"
)

// trackers is the list of Tracker constructors that are available on this system. Tracker implementations should call
// the RegisterTracker function to make themselves available.
var trackers = make(map[string]func() Tracker)

// RegisterTracker makes a Tracker constructor available to clients of this package.
func RegisterTracker(name string, t func() Tracker) {
	if _, exists := trackers[name]; exists {
		log.Fatalf("a tracker already exists with the name %s", name)
	}
	trackers[name] = t
}

// NewTracker returns a new Tracker instance whose type is `name`.
func NewTracker(name string) Tracker {
	if _, exists := trackers[name]; !exists {
		log.Fatalf("no Tracker constructor has been registered with name %s", name)
	}
	return trackers[name]()
}

// Tracker tracks application usage. An implementation that satisfies
// this interface is required for each OS windowing system Thyme
// supports.
type Tracker interface {
	// Snap returns a Snapshot reflecting the currently in-use windows
	// at the current time.
	Snap() (*Snapshot, error)

	// Deps returns a string listing the dependencies that still need
	// to be installed with instructions for how to install them.
	Deps() string
}

// Stream represents all the sampling data gathered by Thyme.
type Stream struct {
	// Snapshots is a list of window snapshots ordered by time.
	Snapshots []*Snapshot
}

// Print returns a pretty-printed representation of the snapshot.
func (s Stream) Print() string {
	var b bytes.Buffer
	for _, snap := range s.Snapshots {
		fmt.Fprintf(&b, "%s", snap.Print())
	}
	return string(b.Bytes())
}

// Snapshot represents the current state of all in-use application
// windows at a moment in time.
type Snapshot struct {
	Time    time.Time
	Windows []*Window
	Active  int64
	Visible []int64
}

// Print returns a pretty-printed representation of the snapshot.
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

// Window represents an application window.
type Window struct {
	// ID is the numerical identifier of the window.
	ID int64

	// Desktop is the numerical identifier of the desktop the
	// window belongs to.  Equal to -1 if the window is sticky.
	Desktop int64

	// Name is the display name of the window (typically what the
	// windowing system shows in the top bar of the window).
	Name string
}

// systemNames is a set of blacklisted window names that are known to
// be used by system windows that aren't visible to the user.
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

// IsSticky returns true if the window is a sticky window (i.e.
// present on all desktops)
func (w *Window) IsSticky() bool {
	return w.Desktop == -1
}

// IsOnDesktop returns true if the window is present on the specified
// desktop
func (w *Window) IsOnDesktop(desktop int64) bool {
	return w.IsSticky() || w.Desktop == desktop
}

const WindowTitleSeparator = " - "

// Info returns more structured metadata about a window. The metadata
// is extracted using heuristics.
//
// Assumptions:
//     1) Most windows use " - " to separate their window names from their content
//     2) Most windows use the " - " with the application name at the end.
//     3) The few programs that reverse this convention only reverse it.
func (w *Window) Info() *Winfo {
	fields := strings.Split(w.Name, WindowTitleSeparator)
	if len(fields) > 1 {
		first := strings.TrimSpace(fields[0])
		last := strings.TrimSpace(fields[len(fields)-1])
		// Special Cases
		if last == "Google Chrome" {
			return &Winfo{
				App:    "Google Chrome",
				SubApp: strings.TrimSpace(fields[len(fields)-2]),
				Title:  strings.Join(fields[0:len(fields)-2], WindowTitleSeparator),
			}
		} else if first == "Slack" {
			return &Winfo{
				App:    "Slack",
				SubApp: strings.TrimSpace(strings.Join(fields[1:], WindowTitleSeparator)),
			}
		}

		// Default Case
		return &Winfo{
			App:   last,
			Title: strings.Join(fields[:len(fields)-1], WindowTitleSeparator),
		}
	}

	return &Winfo{
		Title: w.Name,
	}
}

// Winfo is structured metadata info about a window.
type Winfo struct {
	// App is the application that controls the window.
	App string

	// SubApp is the sub-application that controls the window. An
	// example is a web app (e.g., Sourcegraph) that runs
	// inside a Chrome tab. In this case, the App field would be
	// "Google Chrome" and the SubApp field would be "Sourcegraph".
	SubApp string

	// Title is the title of the window after the App and SubApp name
	// have been stripped.
	Title string
}

// Print returns a pretty-printed representation of the snapshot.
func (w Winfo) Print() string {
	return fmt.Sprintf("[%s|%s|%s]", w.App, w.SubApp, w.Title)
}
