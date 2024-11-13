//go:build !debug
// +build !debug

package utils

func DebugLog(format string, v ...interface{}) {
	// Do nothing
}
