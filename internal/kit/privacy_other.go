//go:build !windows

package kit

import "io/fs"

func protectHistory(data []byte) ([]byte, error)            { return data, nil }
func unprotectHistory(data []byte) ([]byte, error)          { return data, nil }
func privateMode(mode fs.FileMode, wanted fs.FileMode) bool { return mode.Perm() == wanted }
func sameMode(actual, wanted fs.FileMode) bool              { return actual == wanted }
