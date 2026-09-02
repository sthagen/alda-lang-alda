//go:build windows

package system

import "syscall"

// detachedProcess corresponds to the Win32 DETACHED_PROCESS flag, documented
// at https://learn.microsoft.com/windows/win32/procthread/process-creation-flags
//
// It isn't defined in the standard syscall package on Windows, so we define
// it ourselves rather than pull in golang.org/x/sys/windows as a direct
// dependency for a single constant.
const detachedProcess = 0x00000008

// sysProcAttr returns the syscall.SysProcAttr to use when spawning player
// processes.
//
// On Windows, child processes are attached to their parent's console by
// default. When that console is closed, Windows sends a close event to every
// process still attached to it, terminating the player along with the
// terminal. DETACHED_PROCESS spawns the player with no console of its own,
// so it isn't attached to the parent's console and survives the terminal
// closing — matching the behavior on Unix, where the player is spawned in
// its own process group and detaches from the controlling terminal.
func sysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		CreationFlags: detachedProcess,
	}
}
