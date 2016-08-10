package thyme

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type DarwinTracker struct{}

var _ Tracker = (*DarwinTracker)(nil)

func NewDarwinTracker() Tracker {
	return &DarwinTracker{}
}

// allWindowsScript fetches the windows of all scriptable applications.  It
// iterates through each application process known to System Events and attempts
// to script the application with the same name as the application process. If
// such an application exists and is scriptable, it prints the name of every
// window in the application. Otherwise, it just prints the name of every
// visible window in the application. If no visible windows exist, it will just
// print the application name.  (System Events processes only have windows in
// the current desktop/workspace.)
const (
	allWindowsScript = `
tell application "System Events"
  set listOfProcesses to (every application process where background only is false)
end tell

repeat with proc in listOfProcesses
  set procName to (name of proc)
  set procID to (id of proc)
  log "PROCESS " & procID & ":" & procName
  -- Attempt to list windows if the process is scriptable
  try
    tell application procName
      repeat with i from 1 to (count windows)
        log "WINDOW " & (id of window i) & ":" & (name of window i) as string
      end repeat
    end tell
  end try
end repeat
`
	activeWindowsScript = `
tell application "System Events"
     set proc to (first application process whose frontmost is true)
end tell

set procName to (name of proc)
try
  tell application procName
     log "WINDOW " & (id of window 1) & ":" & (name of window 1)
  end tell
on error e
  log "WINDOW " & (id of proc) & ":" & (name of first window of proc)
end try
`
	// visibleWindowsScript generates a mapping from process to windows in the
	// current desktop (note: this is slightly different than the behavior of
	// the previous two scripts, where an empty windows list for a process
	// should NOT imply that there is one window named after the process.
	// Furthermore, the window IDs are not valid in this script (only the window
	// name is valid).
	visibleWindowsScript = `
tell application "System Events"
  repeat with proc in (every process whose visible is true)
    set procName to (name of proc)
    set procID to (id of proc)
    log "PROCESS " & procID & ":" & procName
    tell proc
      repeat with i from 1 to (count windows)
        log "WINDOW -1:" & (name of window i) as string
      end repeat
    end tell
  end repeat
end tell
`
)

func (t *DarwinTracker) Deps() string {
	// TODO: mention requirement of enabling privileged Accessibility tools in preferences
	return ""
}

func (t *DarwinTracker) Snap() (*Snapshot, error) {
	var allWindows []*Window
	var allProcWins map[process][]*Window
	{
		procWins, err := runAS(allWindowsScript)
		if err != nil {
			return nil, err
		}
		for proc, wins := range procWins {
			if len(wins) == 0 {
				allWindows = append(allWindows, &Window{ID: proc.id, Name: proc.name})
			} else {
				allWindows = append(allWindows, wins...)
			}
		}
		allProcWins = procWins
	}

	var active int64
	{
		procWins, err := runAS(activeWindowsScript)
		if err != nil {
			return nil, err
		}
		if len(procWins) > 1 {
			return nil, fmt.Errorf("found more than one active process: %+v", procWins)
		}
		for proc, wins := range procWins {
			if len(wins) == 0 {
				active = proc.id
			} else if len(wins) == 1 {
				active = wins[0].ID
			} else {
				return nil, fmt.Errorf("found more than one active window: %+v", wins)
			}
		}
	}

	var visible []int64
	{
		procWins, err := runAS(visibleWindowsScript)
		if err != nil {
			return nil, err
		}
		for proc, wins := range procWins {
			allWins := allProcWins[proc]
			for _, visWin := range wins {
				if len(allWins) == 0 {
					visible = append(visible, proc.id)
				} else {
					found := false
					for _, win := range allWins {
						if win.Name == visWin.Name {
							visible = append(visible, win.ID)
							found = true
							break
						}
					}
					if !found {
						log.Printf("warning: window ID not found for visible window %q", visWin.Name)
					}
				}
			}
		}
	}

	return &Snapshot{
		Time:    time.Now(),
		Windows: allWindows,
		Active:  active,
		Visible: visible,
	}, nil
}

// process is the {name, id} of a process
type process struct {
	name string
	id   int64
}

// runAS runs script as AppleScript and parses the output into a map of
// processes to windows.
func runAS(script string) (map[process][]*Window, error) {
	cmd := exec.Command("osascript")
	cmd.Stdin = bytes.NewBuffer([]byte(script))
	b, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("AppleScript error: %s, output was:\n%s", err, string(b))
	}
	return parseASOutput(string(b))
}

// parseASOutput parses the output of the AppleScript snippets used to extract window information.
func parseASOutput(out string) (map[process][]*Window, error) {
	proc := process{}
	procWins := make(map[process][]*Window)
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "PROCESS ") {
			c := strings.Index(line, ":")
			procID, err := strconv.ParseInt(line[len("PROCESS "):c], 10, 0)
			if err != nil {
				return nil, err
			}
			proc = process{line[c+1:], procID}
			procWins[proc] = nil
		} else if strings.HasPrefix(line, "WINDOW ") {
			c := strings.Index(line, ":")
			win := line[c+1:]
			winID, err := strconv.ParseInt(line[len("WINDOW "):c], 10, 0)
			if err != nil {
				return nil, err
			}
			procWins[proc] = append(procWins[proc],
				&Window{ID: winID, Name: fmt.Sprintf("%s - %s", win, proc.name)},
			)
		}
	}
	return procWins, nil
}
