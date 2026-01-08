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

package main

import (
	"fmt"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/effector"
)

func main() {
// Create an enforcer with a rate limit model
e, err := casbin.NewEnforcer("../rate_limit_model.conf", "../rate_limit_policy.csv")
if err != nil {
panic(err)
}

// Set the rate limit effector
// This is required to enable rate limiting functionality
rateLimitEft := effector.NewRateLimitEffector()
e.SetEffector(rateLimitEft)

fmt.Println("Rate Limiting Example")
fmt.Println("======================")
fmt.Println("Policy: rate_limit(3, second, allow, sub)")
fmt.Println("This means: Allow at most 3 requests per second, per subject")
fmt.Println()

// Alice tries to access data1 with read permission
// The rate limit is 3 per second, so the first 3 should succeed
for i := 1; i <= 5; i++ {
ok, err := e.Enforce("alice", "data1", "read")
if err != nil {
fmt.Printf("Request %d error: %v\n", i, err)
continue
}
if ok {
fmt.Printf("Request %d: ✓ Allowed\n", i)
} else {
fmt.Printf("Request %d: ✗ Denied (rate limit exceeded)\n", i)
}
}

fmt.Println()
fmt.Println("Bob has a separate rate limit bucket:")

// Bob should have a separate rate limit bucket
for i := 1; i <= 3; i++ {
ok, err := e.Enforce("bob", "data1", "read")
if err != nil {
fmt.Printf("Bob's request %d error: %v\n", i, err)
continue
}
if ok {
fmt.Printf("Bob's request %d: ✓ Allowed\n", i)
} else {
fmt.Printf("Bob's request %d: ✗ Denied\n", i)
}
}
}
