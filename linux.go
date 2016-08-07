package thyme

import (
	"os/exec"
	"strconv"
	"strings"
)

type LinuxTracker struct{}

func NewLinuxTracker() Tracker {
	return &LinuxTracker{}
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
			windows = append(windows, &Window{
				ID:   id,
				Name: name,
			})
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

	return &Snapshot{Windows: windows, Active: active}, nil
}
