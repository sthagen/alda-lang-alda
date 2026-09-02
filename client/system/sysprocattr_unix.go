//go:build !windows

package system

import "syscall"

// sysProcAttr returns the syscall.SysProcAttr to use when spawning player
// processes.
//
// We set Setpgid so that each player process runs in its own process group.
// Otherwise, player processes would inherit the client's process group and be
// killed by terminal-generated signals (e.g. Ctrl+C in a REPL server running in
// server-only mode, where there is no interactive client to put the terminal in
// raw mode).
func sysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true}
}
