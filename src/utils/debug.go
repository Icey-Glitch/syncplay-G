//go:build debug
// +build debug

package utils

import "fmt"

func DebugLog(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}
