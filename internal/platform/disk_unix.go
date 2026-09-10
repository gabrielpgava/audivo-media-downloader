//go:build !windows

package platform

import (
	"path/filepath"

	"golang.org/x/sys/unix"
)

func FreeBytes(path string) (uint64, error) {
	var stat unix.Statfs_t
	if err := unix.Statfs(filepath.Clean(path), &stat); err != nil {
		return 0, err
	}
	return uint64(stat.Bavail) * uint64(stat.Bsize), nil
}
