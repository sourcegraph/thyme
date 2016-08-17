package thyme

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// LinuxTracker tracks application usage on Linux via a few standard command-line utilities.
type LinuxTracker struct{}

var _ Tracker = (*LinuxTracker)(nil)

func NewLinuxTracker() Tracker {
	return &LinuxTracker{}
}

func (t *LinuxTracker) Deps() string {
	return `Install the following command-line utilities via your package manager (e.g., apt) of choice:
* xdpyinfo
* xwininfo
* xdotool
* wmctrl
`
}

func (t *LinuxTracker) Snap() (*Snapshot, error) {
	var windows []*Window
	{
		out, err := exec.Command("wmctrl", "-l").Output()
		if err != nil {
			return nil, err
		}
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) < 4 {
				continue
			}
			id_, name := fields[0], strings.Join(fields[3:], " ")
			id, err := strconv.ParseInt(id_, 0, 64)
			if err != nil {
				return nil, err
			}
			w := Window{ID: id, Name: name}
			if !w.IsSystem() {
				windows = append(windows, &w)
			}
		}
	}

	var visible []int64
	{
		for _, window := range windows {
			out, err := exec.Command("xwininfo", "-id", fmt.Sprintf("%d", window.ID), "-stats").Output()
			if err != nil {
				return nil, err
			}
			if vis.Match(out) {
				visible = append(visible, window.ID)
			}
		}
	}

	var active int64
	{
		out, err := exec.Command("xdotool", "getactivewindow").Output()
		if err != nil {
			return nil, err
		}
		id, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
		if err != nil {
			return nil, err
		}
		active = id
	}

	return &Snapshot{Windows: windows, Active: active, Visible: visible, Time: time.Now()}, nil
}

// isVisible checks if the window is visible in the current viewport.
// x and y are assumed to be relative to the current viewport (i.e.,
// (0, 0) is the coordinate of the top-left corner of the current
// viewport.
func isVisible(x, y, w, h, viewHeight, viewWidth int) bool {
	return (0 <= x && x < viewWidth && 0 <= y && y < viewHeight) ||
		(0 <= x+w && x+w < viewWidth && 0 <= y && y < viewHeight) ||
		(0 <= x && x < viewWidth && 0 <= y+h && y+h < viewHeight) ||
		(0 <= x+w && x+w < viewWidth && 0 <= y+h && y+h < viewHeight)
}

var (
	vis = regexp.MustCompile(`Map State:\s+IsViewable`)
)

// parseWinDim parses window dimension info from the output of `xwininfo`
func parseWinDim(rx *regexp.Regexp, out string, varname string) (int, error) {
	if matches := rx.FindStringSubmatch(out); len(matches) == 2 {
		n, err := strconv.Atoi(matches[1])
		if err != nil {
			return 0, err
		}
		return n, nil
	} else {
		return 0, fmt.Errorf("could not parse window %s from output %s", varname, out)
	}

}
