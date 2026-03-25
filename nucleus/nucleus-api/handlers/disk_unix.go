//go:build !windows

package handlers

import (
	"nucleus-api/models"
	"syscall"
)

func getDiskUsage() models.DiskUsage {
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err != nil {
		return models.DiskUsage{}
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	used := total - free
	percent := 0.0
	if total > 0 {
		percent = float64(used) / float64(total) * 100
	}

	return models.DiskUsage{
		Total:     total,
		Used:      used,
		Available: free,
		Percent:   percent,
	}
}
