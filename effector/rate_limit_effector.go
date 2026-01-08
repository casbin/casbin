// Copyright 2026 The casbin Authors. All Rights Reserved.
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

package effector

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RateLimitBucket holds the state for a rate limit bucket
type RateLimitBucket struct {
	count     int
	windowEnd time.Time
}

// RateLimitEffector is an effector that implements rate limiting
type RateLimitEffector struct {
	mu             sync.RWMutex
	buckets        map[string]*RateLimitBucket
	requestContext map[string]string // stores current request context (sub, obj, act)
}

// NewRateLimitEffector creates a new RateLimitEffector
func NewRateLimitEffector() *RateLimitEffector {
	return &RateLimitEffector{
		buckets:        make(map[string]*RateLimitBucket),
		requestContext: make(map[string]string),
	}
}

var rateLimitRegex = regexp.MustCompile(`rate_limit\((\d+),\s*(\w+),\s*(\w+),\s*(\w+)\)`)

// parseRateLimitExpr parses a rate_limit expression
// Format: rate_limit(max, unit, count_type, bucket)
func parseRateLimitExpr(expr string) (max int, unit string, countType string, bucket string, err error) {
	matches := rateLimitRegex.FindStringSubmatch(expr)
	if matches == nil || len(matches) != 5 {
		return 0, "", "", "", fmt.Errorf("invalid rate_limit expression: %s", expr)
	}

	max, err = strconv.Atoi(matches[1])
	if err != nil {
		return 0, "", "", "", fmt.Errorf("invalid max value: %s", matches[1])
	}

	unit = matches[2]
	countType = matches[3]
	bucket = matches[4]

	// Validate unit
	validUnits := map[string]bool{"second": true, "minute": true, "hour": true, "day": true}
	if !validUnits[unit] {
		return 0, "", "", "", fmt.Errorf("invalid unit: %s (must be second, minute, hour, or day)", unit)
	}

	// Validate count_type
	if countType != "allow" && countType != "all" {
		return 0, "", "", "", fmt.Errorf("invalid count_type: %s (must be allow or all)", countType)
	}

	// Validate bucket
	validBuckets := map[string]bool{"all": true, "sub": true, "obj": true, "act": true}
	if !validBuckets[bucket] {
		return 0, "", "", "", fmt.Errorf("invalid bucket: %s (must be all, sub, obj, or act)", bucket)
	}

	return max, unit, countType, bucket, nil
}

// getWindowDuration returns the duration for a time unit
func getWindowDuration(unit string) time.Duration {
	switch unit {
	case "second":
		return time.Second
	case "minute":
		return time.Minute
	case "hour":
		return time.Hour
	case "day":
		return 24 * time.Hour
	default:
		return time.Minute
	}
}

// MergeEffects implements the Effector interface with rate limiting
func (e *RateLimitEffector) MergeEffects(expr string, effects []Effect, matches []float64, policyIndex int, policyLength int) (Effect, int, error) {
	// Check if this is a rate_limit expression
	if !strings.Contains(expr, "rate_limit") {
		return Deny, -1, errors.New("RateLimitEffector requires rate_limit expression")
	}

	max, unit, countType, bucketType, err := parseRateLimitExpr(expr)
	if err != nil {
		return Deny, -1, err
	}

	// For rate limiting, we need to check all matches first to determine the base result
	var baseEffect Effect = Indeterminate
	var explainIndex int = -1

	// Find the first matching policy
	for i := 0; i < policyLength; i++ {
		if matches[i] != 0 {
			if effects[i] == Allow {
				baseEffect = Allow
				explainIndex = i
				break
			} else if effects[i] == Deny {
				baseEffect = Deny
				explainIndex = i
				break
			}
		}
	}

	// Count this request if needed
	shouldCount := false
	if countType == "all" {
		// Count all requests
		shouldCount = true
	} else if countType == "allow" && baseEffect == Allow {
		// Only count allowed requests
		shouldCount = true
	}

	// Check and update rate limit
	if shouldCount {
		now := time.Now()
		windowDuration := getWindowDuration(unit)

		e.mu.Lock()
		defer e.mu.Unlock()

		// Generate bucket key inside lock to avoid race condition
		bucketKey := e.generateBucketKeyLocked(bucketType)

		bucket, exists := e.buckets[bucketKey]
		if !exists || now.After(bucket.windowEnd) {
			// Create new window
			e.buckets[bucketKey] = &RateLimitBucket{
				count:     1,
				windowEnd: now.Add(windowDuration),
			}
		} else {
			// Increment counter in current window
			bucket.count++
			if bucket.count > max {
				// Rate limit exceeded
				return Deny, -1, nil
			}
		}
	}

	return baseEffect, explainIndex, nil
}

// SetRequestContext sets the request context for bucket key generation
// This should be called before MergeEffects to provide request context
func (e *RateLimitEffector) SetRequestContext(sub, obj, act string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.requestContext["sub"] = sub
	e.requestContext["obj"] = obj
	e.requestContext["act"] = act
}

// generateBucketKeyLocked generates a bucket key based on the bucket type and request context
// IMPORTANT: Must be called with e.mu held to avoid race conditions
func (e *RateLimitEffector) generateBucketKeyLocked(bucketType string) string {
	switch bucketType {
	case "all":
		return "bucket:all"
	case "sub":
		if sub, ok := e.requestContext["sub"]; ok {
			return fmt.Sprintf("bucket:sub:%s", sub)
		}
		return "bucket:sub:unknown"
	case "obj":
		if obj, ok := e.requestContext["obj"]; ok {
			return fmt.Sprintf("bucket:obj:%s", obj)
		}
		return "bucket:obj:unknown"
	case "act":
		if act, ok := e.requestContext["act"]; ok {
			return fmt.Sprintf("bucket:act:%s", act)
		}
		return "bucket:act:unknown"
	default:
		return "bucket:unknown"
	}
}

// GetBucketState returns the current state of a bucket (for testing)
func (e *RateLimitEffector) GetBucketState(key string) (count int, windowEnd time.Time, exists bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	bucket, exists := e.buckets[key]
	if !exists {
		return 0, time.Time{}, false
	}
	return bucket.count, bucket.windowEnd, true
}

// ResetBuckets clears all buckets (for testing)
func (e *RateLimitEffector) ResetBuckets() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.buckets = make(map[string]*RateLimitBucket)
}
