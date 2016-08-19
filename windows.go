// +build windows

package thyme

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

func init() {
	RegisterTracker("windows", NewWindowsTracker)
}

// WindowsTracker tracks application usage using the "EnumWindows" win32 API. Windows is very liberal
// in what it calls a Window, so one Window (or application) may return multiple times with the same
// process ID, but different window titles. Which title is used is a matter of chance.
type WindowsTracker struct{}

var _ Tracker = (*WindowsTracker)(nil)

func NewWindowsTracker() Tracker {
	return &WindowsTracker{}
}

var windowsIgnoreList []string = []string{
	"Default IME",
	"MSCTFIME UI",
}

var (
	user                         = syscall.NewLazyDLL("user32.dll")
	procGetForegroundWindow      = user.NewProc("GetForegroundWindow")
	procGetWindowText            = user.NewProc("GetWindowTextW")
	procGetWindowTextLengthW     = user.NewProc("GetWindowTextLengthW")
	procEnumWindows              = user.NewProc("EnumWindows")
	procIsWindow                 = user.NewProc("IsWindow")
	procIsWindowVisible          = user.NewProc("IsWindowVisible")
	procGetWindowThreadProcessId = user.NewProc("GetWindowThreadProcessId")
)

func (t *WindowsTracker) Deps() string {
	return "Nothing, Ready to Go!"
}

// getWindowTitle returns a title of a window of the provided system window handle
func getWindowTitle(window uintptr) string {
	textLength, _, _ := procGetWindowTextLengthW.Call(uintptr(window))
	textLength += 1
	titleBuffer := make([]uint16, textLength)
	procGetWindowText.Call(uintptr(window), uintptr(unsafe.Pointer(&titleBuffer[0])), textLength)
	return syscall.UTF16ToString(titleBuffer)
}

// getWindowID returns the process (thread) id that created the window. Multiple windows can share
// the same process id.
func getWindowID(window uintptr) int64 {
	id, _, _ := procGetWindowThreadProcessId.Call(window)
	return int64(id)
}

// windowsIgnore will return true for titles of windows that are likely internal to windows itself
// and not the applications we care to monitor.
func windowsIgnore(title string) bool {
	if title == "" {
		return true
	}
	for _, ignore := range windowsIgnoreList {
		if ignore == title {
			return true
		}
	}
	return false
}

func (t *WindowsTracker) Snap() (snap *Snapshot, err error) {
	var allWindows []*Window
	var visible []int64
	var active int64

	var cbId uintptr = 888

	activeWindow, _, _ := procGetForegroundWindow.Call()
	activeTitle := getWindowTitle(activeWindow)

	cb := syscall.NewCallback(func(hwnd syscall.Handle, lparam uintptr) uintptr {
		if lparam != cbId {
			err = fmt.Errorf("lparam does not match what callback expected; received (%d), expected (%d)", lparam, cbId)
			return 0
		}
		b, _, _ := procIsWindow.Call(uintptr(hwnd))
		if b != 0 {
			currentTitle := getWindowTitle(uintptr(hwnd))
			currentId := getWindowID(uintptr(hwnd))
			// Skip windows that are in a process where we already have a visible window
			for _, visibleId := range visible {
				if currentId == visibleId {
					return 1
				}
			}
			if !windowsIgnore(currentTitle) {
				if activeTitle == currentTitle {
					active = currentId
				}
				v, _, _ := procIsWindowVisible.Call(uintptr(hwnd))
				if v != 0 {
					visible = append(visible, currentId)
				}
				allWindows = append(allWindows, &Window{ID: currentId, Name: currentTitle})
			}
		}
		return 1 // continue enumeration
	})

	procEnumWindows.Call(cb, cbId)

	return &Snapshot{
		Time:    time.Now(),
		Windows: allWindows,
		Active:  active,
		Visible: visible,
	}, err
}
