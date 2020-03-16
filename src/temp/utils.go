package mensa

import (
	"net"
	"os"
	"runtime"
	"syscall"
)

// IsAddressInUse check error
func IsAddressInUse(err error) bool {
	errOpError, ok := err.(*net.OpError)
	if !ok {
		return false
	}
	errSyscallError, ok := errOpError.Err.(*os.SyscallError)
	if !ok {
		return false
	}
	errErrno, ok := errSyscallError.Err.(syscall.Errno)
	if !ok {
		return false
	}
	if errErrno == syscall.EADDRINUSE {
		return true
	}
	const WSAEADDRINUSE = 10048
	return runtime.GOOS == "windows" && errErrno == WSAEADDRINUSE
}
