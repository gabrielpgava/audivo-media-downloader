//go:build windows

package platform

import (
	"path/filepath"

	"golang.org/x/sys/windows"
)

func FreeBytes(path string) (uint64, error) {
	var free, total, totalFree uint64
	root := filepath.VolumeName(filepath.Clean(path)) + `\`
	if err := windows.GetDiskFreeSpaceEx(windows.StringToUTF16Ptr(root), &free, &total, &totalFree); err != nil {
		return 0, err
	}
	return free, nil
}
