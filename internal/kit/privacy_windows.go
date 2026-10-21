package kit

import (
	"errors"
	"io/fs"
	"unsafe"

	"golang.org/x/sys/windows"
)

func cryptHistory(data []byte, encrypt bool) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("histórico vazio")
	}
	input := windows.DataBlob{Size: uint32(len(data)), Data: &data[0]}
	var output windows.DataBlob
	var err error
	if encrypt {
		err = windows.CryptProtectData(&input, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &output)
	} else {
		err = windows.CryptUnprotectData(&input, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &output)
	}
	if err != nil {
		return nil, err
	}
	defer windows.LocalFree(windows.Handle(uintptr(unsafe.Pointer(output.Data))))
	return append([]byte(nil), unsafe.Slice(output.Data, int(output.Size))...), nil
}

func protectHistory(data []byte) ([]byte, error)    { return cryptHistory(data, true) }
func unprotectHistory(data []byte) ([]byte, error)  { return cryptHistory(data, false) }
func privateMode(_ fs.FileMode, _ fs.FileMode) bool { return true }
func sameMode(actual, wanted fs.FileMode) bool      { return actual&0200 == wanted&0200 }
