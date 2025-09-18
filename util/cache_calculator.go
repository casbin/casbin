// Copyright 2024 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"runtime"

	"github.com/shirou/gopsutil/v3/mem"
)

// CalculateDynamicCacheSize dynamically calculates cache size based on real system memory usage.
func CalculateDynamicCacheSize() int {
	// Get real system memory information
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		// If unable to get system memory info, use default value
		return 1000000 // Default 1 million entries
	}

	// Get Go program memory usage
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Set reasonable range limits
	const (
		minCacheSize = 10000    // Minimum cache size (10k entries)
		maxCacheSize = 50000000 // Maximum cache size (50M entries)
	)

	// Calculate system available memory
	systemAvailableMemory := memInfo.Available
	systemTotalMemory := memInfo.Total
	systemUsedMemory := memInfo.Used

	// Calculate system memory usage rate
	systemMemoryUsageRate := float64(systemUsedMemory) / float64(systemTotalMemory)
	systemAvailableRate := float64(systemAvailableMemory) / float64(systemTotalMemory)

	// Detect system pressure level
	pressureLevel := detectSystemPressureWithRealMemory(m, systemMemoryUsageRate, systemAvailableRate)

	// Adjust cache size based on pressure level
	var cacheMemoryLimit uint64
	switch pressureLevel {
	case "low":
		// Low pressure: use 20% of system available memory for cache
		cacheMemoryLimit = systemAvailableMemory / 5
	case "medium":
		// Medium pressure: use 10% of system available memory for cache
		cacheMemoryLimit = systemAvailableMemory / 10
	case "high":
		// High pressure: use 5% of system available memory for cache
		cacheMemoryLimit = systemAvailableMemory / 20
	case "critical":
		// Critical pressure: use minimum cache
		return minCacheSize
	default:
		// Default: use 10% of system available memory for cache
		cacheMemoryLimit = systemAvailableMemory / 10
	}

	// Each cache entry approximately takes 67 bytes (key + value + overhead)
	// Conservative estimate: 100 bytes per entry
	bytesPerEntry := uint64(100)

	// Calculate maximum cache entries
	maxEntries := cacheMemoryLimit / bytesPerEntry

	// Limit within reasonable range
	if maxEntries < minCacheSize {
		return minCacheSize
	}
	if maxEntries > maxCacheSize {
		return maxCacheSize
	}

	return int(maxEntries)
}

// detectSystemPressureWithRealMemory detects system pressure level based on real system memory.
func detectSystemPressureWithRealMemory(m runtime.MemStats, systemMemoryUsageRate, systemAvailableRate float64) string {
	// Calculate GC pressure
	gcPressure := float64(m.NumGC) / float64(m.NumForcedGC+1) // Avoid division by zero

	// Pressure detection thresholds
	const (
		criticalSystemMemoryRate = 0.90 // System memory usage > 90% is critical pressure
		highSystemMemoryRate     = 0.80 // System memory usage > 80% is high pressure
		mediumSystemMemoryRate   = 0.70 // System memory usage > 70% is medium pressure

		criticalSystemAvailableRate = 0.10 // System available memory < 10% is critical pressure
		highSystemAvailableRate     = 0.20 // System available memory < 20% is high pressure
		mediumSystemAvailableRate   = 0.30 // System available memory < 30% is medium pressure

		highGCPressure   = 10.0 // GC pressure > 10 is high pressure
		mediumGCPressure = 5.0  // GC pressure > 5 is medium pressure
	)

	// Detect critical pressure
	if systemMemoryUsageRate > criticalSystemMemoryRate ||
		systemAvailableRate < criticalSystemAvailableRate {
		return "critical"
	}

	// Detect high pressure
	if systemMemoryUsageRate > highSystemMemoryRate ||
		systemAvailableRate < highSystemAvailableRate ||
		gcPressure > highGCPressure {
		return "high"
	}

	// Detect medium pressure
	if systemMemoryUsageRate > mediumSystemMemoryRate ||
		systemAvailableRate < mediumSystemAvailableRate ||
		gcPressure > mediumGCPressure {
		return "medium"
	}

	// Low pressure
	return "low"
}
