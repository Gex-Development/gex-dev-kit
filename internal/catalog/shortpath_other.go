//go:build !windows

package catalog

import "errors"

func shortPath(string) (string, error) {
	return "", errors.New("caminho curto 8.3 só existe no Windows")
}
