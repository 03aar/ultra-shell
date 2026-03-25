//go:build windows

package handlers

import (
	"nucleus-api/models"
	"os/exec"
	"strconv"
	"strings"
)

func getDiskUsage() models.DiskUsage {
	// Use wmic on Windows to get disk usage
	out, err := exec.Command("wmic", "logicaldisk", "where", "DeviceID='C:'",
		"get", "Size,FreeSpace", "/format:csv").Output()
	if err != nil {
		return models.DiskUsage{}
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return models.DiskUsage{}
	}

	// Parse CSV: Node,FreeSpace,Size
	fields := strings.Split(strings.TrimSpace(lines[len(lines)-1]), ",")
	if len(fields) < 3 {
		return models.DiskUsage{}
	}

	free, _ := strconv.ParseUint(strings.TrimSpace(fields[1]), 10, 64)
	total, _ := strconv.ParseUint(strings.TrimSpace(fields[2]), 10, 64)
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
