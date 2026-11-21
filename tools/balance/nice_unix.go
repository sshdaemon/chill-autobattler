//go:build unix

package main

import "syscall"

// lowerPriority runs the lab at nice 10, so anything interactive goes first.
func lowerPriority() { _ = syscall.Setpriority(syscall.PRIO_PROCESS, 0, 10) }
