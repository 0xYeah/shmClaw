//go:build windows

package shm

import (
	"errors"
	"os"
)

func mmapFile(file *os.File, size int64) ([]byte, error) {
	return nil, errors.New("mmap not implemented for windows")
}

func munmapFile(data []byte) error {
	return errors.New("munmap not implemented for windows")
}
