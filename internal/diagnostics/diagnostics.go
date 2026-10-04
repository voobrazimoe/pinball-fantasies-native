// Package diagnostics controls optional host chatter. Errors bypass this gate.
package diagnostics

import (
	"fmt"
	"os"
)

func Enabled() bool { return os.Getenv("PF_DIAGNOSTICS") == "1" }
func Printf(format string, args ...any) {
	if Enabled() {
		fmt.Printf(format, args...)
	}
}
func Println(args ...any) {
	if Enabled() {
		fmt.Println(args...)
	}
}
