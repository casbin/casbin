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
	"testing"
	"time"
)

func TestRateLimitEffectorBasic(t *testing.T) {
	eft := NewRateLimitEffector()

	// Test rate limiting with allow count type
	expr := "rate_limit(2, second, allow, sub)"
	effects := []Effect{Allow}
	matches := []float64{1.0}

	// First request should succeed
	effect, _, err := eft.MergeEffects(expr, effects, matches, 0, 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if effect != Allow {
		t.Errorf("Expected Allow, got %v", effect)
	}

	// Second request should succeed
	effect, _, err = eft.MergeEffects(expr, effects, matches, 0, 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if effect != Allow {
		t.Errorf("Expected Allow, got %v", effect)
	}

	// Third request should be denied (exceeds limit of 2)
	effect, _, err = eft.MergeEffects(expr, effects, matches, 0, 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if effect != Deny {
		t.Errorf("Expected Deny due to rate limit, got %v", effect)
	}
}

func TestRateLimitEffectorWindowReset(t *testing.T) {
	eft := NewRateLimitEffector()

	// Use a short window for testing
	expr := "rate_limit(2, second, allow, sub)"
	effects := []Effect{Allow}
	matches := []float64{1.0}

	// First request
	effect, _, _ := eft.MergeEffects(expr, effects, matches, 0, 1)
	if effect != Allow {
		t.Errorf("Expected Allow, got %v", effect)
	}

	// Second request
	effect, _, _ = eft.MergeEffects(expr, effects, matches, 0, 1)
	if effect != Allow {
		t.Errorf("Expected Allow, got %v", effect)
	}

	// Third request should be denied
	effect, _, _ = eft.MergeEffects(expr, effects, matches, 0, 1)
	if effect != Deny {
		t.Errorf("Expected Deny, got %v", effect)
	}

	// Wait for window to expire
	time.Sleep(1100 * time.Millisecond)

	// After window reset, request should succeed again
	effect, _, _ = eft.MergeEffects(expr, effects, matches, 0, 1)
	if effect != Allow {
		t.Errorf("Expected Allow after window reset, got %v", effect)
	}
}

func TestRateLimitEffectorCountTypeAll(t *testing.T) {
	eft := NewRateLimitEffector()

	// Count all requests, even denied ones
	expr := "rate_limit(2, second, all, sub)"

	// First request - Allow
	effects := []Effect{Allow}
	matches := []float64{1.0}
	effect, _, _ := eft.MergeEffects(expr, effects, matches, 0, 1)
	if effect != Allow {
		t.Errorf("Expected Allow, got %v", effect)
	}

	// Second request - Deny (but still counts)
	effects = []Effect{Deny}
	effect, _, _ = eft.MergeEffects(expr, effects, matches, 0, 1)
	if effect != Deny {
		t.Errorf("Expected Deny from policy, got %v", effect)
	}

	// Third request - should be denied by rate limit even if policy allows
	effects = []Effect{Allow}
	effect, _, _ = eft.MergeEffects(expr, effects, matches, 0, 1)
	if effect != Deny {
		t.Errorf("Expected Deny due to rate limit, got %v", effect)
	}
}

func TestRateLimitEffectorCountTypeAllow(t *testing.T) {
	eft := NewRateLimitEffector()

	// Only count allowed requests
	expr := "rate_limit(2, second, allow, sub)"

	// First request - Allow (counts)
	effects := []Effect{Allow}
	matches := []float64{1.0}
	effect, _, _ := eft.MergeEffects(expr, effects, matches, 0, 1)
	if effect != Allow {
		t.Errorf("Expected Allow, got %v", effect)
	}

	// Second request - Deny (doesn't count)
	effects = []Effect{Deny}
	effect, _, _ = eft.MergeEffects(expr, effects, matches, 0, 1)
	if effect != Deny {
		t.Errorf("Expected Deny from policy, got %v", effect)
	}

	// Third request - Allow (counts, should be allowed as we only counted 1 so far)
	effects = []Effect{Allow}
	effect, _, _ = eft.MergeEffects(expr, effects, matches, 0, 1)
	if effect != Allow {
		t.Errorf("Expected Allow, got %v", effect)
	}

	// Fourth request - should be denied by rate limit
	effect, _, _ = eft.MergeEffects(expr, effects, matches, 0, 1)
	if effect != Deny {
		t.Errorf("Expected Deny due to rate limit, got %v", effect)
	}
}

func TestRateLimitEffectorDifferentTimeUnits(t *testing.T) {
	testCases := []struct {
		unit     string
		duration time.Duration
	}{
		{"second", time.Second},
		{"minute", time.Minute},
		{"hour", time.Hour},
		{"day", 24 * time.Hour},
	}

	for _, tc := range testCases {
		t.Run(tc.unit, func(t *testing.T) {
			eft := NewRateLimitEffector()
			expr := "rate_limit(1, " + tc.unit + ", allow, sub)"
			effects := []Effect{Allow}
			matches := []float64{1.0}

			// First request should succeed
			effect, _, err := eft.MergeEffects(expr, effects, matches, 0, 1)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if effect != Allow {
				t.Errorf("Expected Allow, got %v", effect)
			}

			// Second request should be denied
			effect, _, err = eft.MergeEffects(expr, effects, matches, 0, 1)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if effect != Deny {
				t.Errorf("Expected Deny, got %v", effect)
			}
		})
	}
}

func TestRateLimitEffectorInvalidExpressions(t *testing.T) {
	eft := NewRateLimitEffector()
	effects := []Effect{Allow}
	matches := []float64{1.0}

	testCases := []struct {
		name string
		expr string
	}{
		{"invalid format", "rate_limit(10)"},
		{"invalid max", "rate_limit(abc, second, allow, sub)"},
		{"invalid unit", "rate_limit(10, week, allow, sub)"},
		{"invalid count_type", "rate_limit(10, second, maybe, sub)"},
		{"invalid bucket", "rate_limit(10, second, allow, user)"},
		{"not rate_limit", "some(where (p.eft == allow))"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := eft.MergeEffects(tc.expr, effects, matches, 0, 1)
			if err == nil {
				t.Errorf("Expected error for invalid expression: %s", tc.expr)
			}
		})
	}
}

func TestRateLimitEffectorNoMatch(t *testing.T) {
	eft := NewRateLimitEffector()

	expr := "rate_limit(2, second, allow, sub)"
	effects := []Effect{Allow}
	matches := []float64{0.0} // No match

	// Should return Indeterminate when no policy matches
	effect, explainIndex, err := eft.MergeEffects(expr, effects, matches, 0, 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if effect != Indeterminate {
		t.Errorf("Expected Indeterminate, got %v", effect)
	}
	if explainIndex != -1 {
		t.Errorf("Expected explainIndex -1, got %d", explainIndex)
	}
}

func TestRateLimitEffectorMultiplePolicies(t *testing.T) {
	eft := NewRateLimitEffector()

	expr := "rate_limit(2, second, allow, sub)"
	effects := []Effect{Deny, Allow, Allow}
	matches := []float64{0.0, 1.0, 0.0} // Only second policy matches

	// First request with second policy matching
	effect, explainIndex, err := eft.MergeEffects(expr, effects, matches, 1, 3)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if effect != Allow {
		t.Errorf("Expected Allow, got %v", effect)
	}
	if explainIndex != 1 {
		t.Errorf("Expected explainIndex 1, got %d", explainIndex)
	}

	// Second request
	effect, _, _ = eft.MergeEffects(expr, effects, matches, 1, 3)
	if effect != Allow {
		t.Errorf("Expected Allow, got %v", effect)
	}

	// Third request should be denied by rate limit
	effect, explainIndex, _ = eft.MergeEffects(expr, effects, matches, 1, 3)
	if effect != Deny {
		t.Errorf("Expected Deny, got %v", effect)
	}
	if explainIndex != -1 {
		t.Errorf("Expected explainIndex -1 for rate limit, got %d", explainIndex)
	}
}

func TestRateLimitEffectorResetBuckets(t *testing.T) {
	eft := NewRateLimitEffector()

	expr := "rate_limit(1, second, allow, sub)"
	effects := []Effect{Allow}
	matches := []float64{1.0}

	// First request
	effect, _, _ := eft.MergeEffects(expr, effects, matches, 0, 1)
	if effect != Allow {
		t.Errorf("Expected Allow, got %v", effect)
	}

	// Second request should be denied
	effect, _, _ = eft.MergeEffects(expr, effects, matches, 0, 1)
	if effect != Deny {
		t.Errorf("Expected Deny, got %v", effect)
	}

	// Reset buckets
	eft.ResetBuckets()

	// After reset, request should succeed
	effect, _, _ = eft.MergeEffects(expr, effects, matches, 0, 1)
	if effect != Allow {
		t.Errorf("Expected Allow after reset, got %v", effect)
	}
}

func TestRateLimitEffectorWithContext(t *testing.T) {
	eft := NewRateLimitEffector()

	expr := "rate_limit(2, second, allow, sub)"
	effects := []Effect{Allow}
	matches := []float64{1.0}

	// Set context for alice
	eft.SetRequestContext("alice", "data1", "read")

	// First two requests from alice should succeed
	effect, _, _ := eft.MergeEffects(expr, effects, matches, 0, 1)
	if effect != Allow {
		t.Errorf("Expected Allow for alice, got %v", effect)
	}

	effect, _, _ = eft.MergeEffects(expr, effects, matches, 0, 1)
	if effect != Allow {
		t.Errorf("Expected Allow for alice, got %v", effect)
	}

	// Third request from alice should be denied
	effect, _, _ = eft.MergeEffects(expr, effects, matches, 0, 1)
	if effect != Deny {
		t.Errorf("Expected Deny for alice, got %v", effect)
	}

	// Switch context to bob - should have separate bucket
	eft.SetRequestContext("bob", "data1", "read")

	// Bob's first request should succeed (separate bucket)
	effect, _, _ = eft.MergeEffects(expr, effects, matches, 0, 1)
	if effect != Allow {
		t.Errorf("Expected Allow for bob, got %v", effect)
	}
}

func TestRateLimitEffectorDifferentBucketTypes(t *testing.T) {
	testCases := []struct {
		name       string
		bucketType string
		context1   []string // [sub, obj, act]
		context2   []string
		sameKey    bool
	}{
		{"bucket by all", "all", []string{"alice", "data1", "read"}, []string{"bob", "data2", "write"}, true},
		{"bucket by sub", "sub", []string{"alice", "data1", "read"}, []string{"alice", "data2", "write"}, true},
		{"bucket by sub different", "sub", []string{"alice", "data1", "read"}, []string{"bob", "data1", "read"}, false},
		{"bucket by obj", "obj", []string{"alice", "data1", "read"}, []string{"bob", "data1", "write"}, true},
		{"bucket by obj different", "obj", []string{"alice", "data1", "read"}, []string{"alice", "data2", "read"}, false},
		{"bucket by act", "act", []string{"alice", "data1", "read"}, []string{"bob", "data2", "read"}, true},
		{"bucket by act different", "act", []string{"alice", "data1", "read"}, []string{"alice", "data1", "write"}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			eft := NewRateLimitEffector()
			expr := "rate_limit(1, second, allow, " + tc.bucketType + ")"
			effects := []Effect{Allow}
			matches := []float64{1.0}

			// First request with context1
			eft.SetRequestContext(tc.context1[0], tc.context1[1], tc.context1[2])
			effect, _, _ := eft.MergeEffects(expr, effects, matches, 0, 1)
			if effect != Allow {
				t.Errorf("Expected Allow for first request, got %v", effect)
			}

			// Second request with context2
			eft.SetRequestContext(tc.context2[0], tc.context2[1], tc.context2[2])
			effect, _, _ = eft.MergeEffects(expr, effects, matches, 0, 1)

			if tc.sameKey {
				// Should be denied because they share the same bucket
				if effect != Deny {
					t.Errorf("Expected Deny (same bucket), got %v", effect)
				}
			} else {
				// Should be allowed because they have different buckets
				if effect != Allow {
					t.Errorf("Expected Allow (different bucket), got %v", effect)
				}
			}
		})
	}
}
