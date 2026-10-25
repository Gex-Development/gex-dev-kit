//go:build !windows

package terminal

import "os"

func terminalEnableVT(*os.File) bool {
	return true
}
