//go:build windows

package catalog

import (
	"errors"

	"golang.org/x/sys/windows"
)

func shortPath(long string) (string, error) {
	name, err := windows.UTF16PtrFromString(long)
	if err != nil {
		return "", err
	}
	size, err := windows.GetShortPathName(name, nil, 0)
	if err != nil {
		return "", err
	}
	if size == 0 {
		return "", errors.New("caminho curto indisponível")
	}
	buffer := make([]uint16, size)
	written, err := windows.GetShortPathName(name, &buffer[0], size)
	if err != nil {
		return "", err
	}
	if written == 0 || written > size {
		return "", errors.New("caminho curto inválido")
	}
	return windows.UTF16ToString(buffer[:written]), nil
}
