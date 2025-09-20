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
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
)

// Cache size calculation cache with TTL.
var (
	cachedCacheSize     int
	cachedCacheSizeTime time.Time
	cacheSizeMu         sync.RWMutex
	cacheSizeTTL        = 10 * time.Second // Cache for 10 seconds
)

// CalculateDynamicCacheSize dynamically calculates cache size based on real system memory usage.
// Uses caching to reduce expensive system calls and avoid frequent recalculation.
func CalculateDynamicCacheSize() int {
	// Check cache first - most calls will hit this fast path
	cacheSizeMu.RLock()
	if time.Since(cachedCacheSizeTime) < cacheSizeTTL {
		result := cachedCacheSize
		cacheSizeMu.RUnlock()
		return result
	}
	cacheSizeMu.RUnlock()

	// Cache expired or not set, calculate new value
	newCacheSize := calculateCacheSizeInternal()

	// Update cache
	cacheSizeMu.Lock()
	cachedCacheSize = newCacheSize
	cachedCacheSizeTime = time.Now()
	cacheSizeMu.Unlock()

	return newCacheSize
}

// calculateCacheSizeInternal performs the actual expensive calculation.
func calculateCacheSizeInternal() int {
	// Get real system memory information
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		// If unable to get system memory info, use default value
		return 1000000 // Default 1 million entries
	}

	// Simplified pressure detection - only use system memory info
	// Avoid expensive GC stats for better performance
	systemMemoryUsageRate := float64(memInfo.Used) / float64(memInfo.Total)
	systemAvailableRate := float64(memInfo.Available) / float64(memInfo.Total)

	// Optimized pressure detection thresholds - increased cache limits for better performance
	var cacheMemoryLimit uint64
	switch {
	case systemAvailableRate < 0.05: // < 5% available (more aggressive threshold)
		return 50000 // Increased minimum cache
	case systemAvailableRate < 0.15: // < 15% available
		cacheMemoryLimit = memInfo.Available / 8 // 12.5% of available (increased from 5%)
	case systemAvailableRate < 0.25: // < 25% available
		cacheMemoryLimit = memInfo.Available / 4 // 25% of available (increased from 10%)
	case systemMemoryUsageRate > 0.85: // > 85% used (more aggressive threshold)
		cacheMemoryLimit = memInfo.Available / 4 // 25% of available (increased from 10%)
	default:
		cacheMemoryLimit = memInfo.Available / 2 // 50% of available (increased from 20%)
	}

	// Set reasonable range limits - increased for better performance
	const (
		minCacheSize = 50000     // Increased minimum cache size (50k entries)
		maxCacheSize = 100000000 // Increased maximum cache size (100M entries)
	)

	// Conservative estimate: 50 bytes per entry
	bytesPerEntry := uint64(50)

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
