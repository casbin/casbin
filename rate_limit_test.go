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

package casbin

import (
	"testing"
	"time"

	"github.com/casbin/casbin/v3/effector"
	"github.com/casbin/casbin/v3/model"
)

func TestRateLimitWithEnforcer(t *testing.T) {
	e, err := NewEnforcer("examples/rate_limit_model.conf", "examples/rate_limit_policy.csv")
	if err != nil {
		t.Fatal(err)
	}

	// Set the rate limit effector
	rateLimitEft := effector.NewRateLimitEffector()
	e.SetEffector(rateLimitEft)

	// Test rate limiting for alice (limit: 3 per second)
	// First 3 requests should succeed
	for i := 0; i < 3; i++ {
		ok, err2 := e.Enforce("alice", "data1", "read")
		if err2 != nil {
			t.Errorf("Request %d failed: %v", i+1, err2)
		}
		if !ok {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 4th request should be denied by rate limiter
	ok, err := e.Enforce("alice", "data1", "read")
	if err != nil {
		t.Errorf("Request 4 failed: %v", err)
	}
	if ok {
		t.Error("Request 4 should be denied by rate limiter")
	}

	// Bob should have separate rate limit bucket
	ok, err = e.Enforce("bob", "data1", "read")
	if err != nil {
		t.Errorf("Bob's request failed: %v", err)
	}
	if !ok {
		t.Error("Bob's request should be allowed (separate bucket)")
	}
}

func TestRateLimitWithEnforcerWindowReset(t *testing.T) {
	e, err := NewEnforcer("examples/rate_limit_model.conf", "examples/rate_limit_policy.csv")
	if err != nil {
		t.Fatal(err)
	}

	// Set the rate limit effector
	rateLimitEft := effector.NewRateLimitEffector()
	e.SetEffector(rateLimitEft)

	// First 3 requests should succeed
	for i := 0; i < 3; i++ {
		ok, _ := e.Enforce("alice", "data1", "read")
		if !ok {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 4th request should be denied
	ok, _ := e.Enforce("alice", "data1", "read")
	if ok {
		t.Error("Request 4 should be denied by rate limiter")
	}

	// Wait for window to reset
	time.Sleep(1100 * time.Millisecond)

	// After window reset, request should succeed
	ok, _ = e.Enforce("alice", "data1", "read")
	if !ok {
		t.Error("Request should be allowed after window reset")
	}
}

func TestRateLimitDifferentBucketTypes(t *testing.T) {
	testCases := []struct {
		name      string
		modelText string
		requests  [][]interface{}
		expected  []bool
	}{
		{
			name: "bucket by all",
			modelText: `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = rate_limit(2, second, allow, all)

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
`,
			requests: [][]interface{}{
				{"alice", "data1", "read"},  // 1st - allowed
				{"bob", "data1", "read"},    // 2nd - allowed (shares bucket with alice)
				{"alice", "data2", "write"}, // 3rd - denied (same bucket)
			},
			expected: []bool{true, true, false},
		},
		{
			name: "bucket by obj",
			modelText: `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = rate_limit(2, second, allow, obj)

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
`,
			requests: [][]interface{}{
				{"alice", "data1", "read"},  // 1st - allowed
				{"bob", "data1", "read"},    // 2nd - allowed (same obj bucket)
				{"alice", "data1", "write"}, // 3rd - denied (same obj bucket)
				{"alice", "data2", "write"}, // 4th - allowed (different obj bucket)
			},
			expected: []bool{true, true, false, true},
		},
		{
			name: "bucket by act",
			modelText: `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = rate_limit(2, second, allow, act)

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
`,
			requests: [][]interface{}{
				{"alice", "data1", "read"},  // 1st - allowed
				{"bob", "data2", "read"},    // 2nd - allowed (same act bucket)
				{"alice", "data1", "read"},  // 3rd - denied (same act bucket)
				{"alice", "data1", "write"}, // 4th - allowed (different act bucket)
			},
			expected: []bool{true, true, false, true},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := model.NewModelFromString(tc.modelText)
			if err != nil {
				t.Fatalf("Failed to create model: %v", err)
			}
			e, err := NewEnforcer(m)
			if err != nil {
				t.Fatalf("Failed to create enforcer: %v", err)
			}

			// Add policies
			e.AddPolicy("alice", "data1", "read")
			e.AddPolicy("alice", "data1", "write")
			e.AddPolicy("alice", "data2", "write")
			e.AddPolicy("bob", "data1", "read")
			e.AddPolicy("bob", "data2", "read")
			e.AddPolicy("bob", "data2", "write")

			// Set the rate limit effector
			rateLimitEft := effector.NewRateLimitEffector()
			e.SetEffector(rateLimitEft)

			// Test requests
			for i, req := range tc.requests {
				ok, err := e.Enforce(req...)
				if err != nil {
					t.Errorf("Request %d failed: %v", i+1, err)
				}
				if ok != tc.expected[i] {
					t.Errorf("Request %d (sub=%v, obj=%v, act=%v): expected %v, got %v",
						i+1, req[0], req[1], req[2], tc.expected[i], ok)
				}
			}
		})
	}
}

func TestRateLimitCountTypeAll(t *testing.T) {
	modelText := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = rate_limit(2, second, all, sub)

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
`
	m, err := model.NewModelFromString(modelText)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	e, err := NewEnforcer(m)
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Add policy only for alice/data1/read
	e.AddPolicy("alice", "data1", "read")

	// Set the rate limit effector
	rateLimitEft := effector.NewRateLimitEffector()
	e.SetEffector(rateLimitEft)

	// First request - matches policy (allowed)
	ok, _ := e.Enforce("alice", "data1", "read")
	if !ok {
		t.Error("First request should be allowed")
	}

	// Second request - doesn't match policy (denied by policy, but counts)
	ok, _ = e.Enforce("alice", "data2", "write")
	if ok {
		t.Error("Second request should be denied by policy")
	}

	// Third request - should be denied by rate limiter even if policy matches
	// (because count_type is "all" and we already counted 2 requests)
	ok, _ = e.Enforce("alice", "data1", "read")
	if ok {
		t.Error("Third request should be denied by rate limiter")
	}
}
