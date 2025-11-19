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

package main

import (
	"fmt"
	"log"

	"github.com/casbin/casbin/v2"
)

func main() {
	// Initialize enforcer with priority fallback model and policies
	e, err := casbin.NewEnforcer(
		"priority_fallback_model.conf",
		"priority_fallback_policy.csv",
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Priority-Based Fallback Policy Example ===")
	fmt.Println()

	// Test Case 1: Direct policy match (priority 10)
	fmt.Println("Case 1: Direct Policy Match (Highest Priority)")
	testEnforce(e, "alice", "data1", "read", "Alice has explicit allow policy at priority 10")
	testEnforce(e, "alice", "data1", "write", "Alice has explicit allow policy at priority 10")
	fmt.Println()

	// Test Case 2: Fallback policy match (priority 100)
	fmt.Println("Case 2: Fallback Policy Match (via Role)")
	testEnforce(e, "alice", "data2", "read", "Alice has no explicit policy for data2, falls back to fallback_admin role at priority 100")
	testEnforce(e, "alice", "data2", "write", "Alice has no explicit policy for data2, falls back to fallback_admin role at priority 100")
	fmt.Println()

	// Test Case 3: Priority override
	fmt.Println("Case 3: Priority Override (Higher Priority Wins)")
	testEnforce(e, "bob", "data2", "write", "Bob has explicit DENY at priority 10, which overrides ALLOW from fallback_admin at priority 100")
	fmt.Println()

	// Test Case 4: Fallback for bob on data1
	fmt.Println("Case 4: Another Fallback Example")
	testEnforce(e, "bob", "data1", "read", "Bob has no explicit policy for data1, falls back to fallback_admin role at priority 100")
	testEnforce(e, "bob", "data1", "write", "Bob has no explicit policy for data1, falls back to fallback_admin role at priority 100")
	fmt.Println()

	// Test Case 5: No matching policy at all
	fmt.Println("Case 5: No Matching Policy (Default Deny)")
	testEnforce(e, "charlie", "data3", "read", "Charlie has no policies and no roles, defaults to deny")
	fmt.Println()

	// Demonstrate dynamic policy addition
	fmt.Println("=== Dynamic Policy Management ===")
	fmt.Println()
	fmt.Println("Adding high priority policy for charlie...")
	_, err = e.AddPolicy("5", "charlie", "data1", "read", "allow")
	if err != nil {
		log.Printf("Error adding policy: %v", err)
	}
	testEnforce(e, "charlie", "data1", "read", "Charlie now has explicit allow policy at priority 5")
	fmt.Println()

	// Show which policy matched using EnforceEx
	fmt.Println("=== Detailed Explanation (EnforceEx) ===")
	fmt.Println()
	allowed, explain, _ := e.EnforceEx("alice", "data1", "read")
	fmt.Printf("Request: alice, data1, read\n")
	fmt.Printf("Result: %v\n", allowed)
	fmt.Printf("Matched Policy: %v\n\n", explain)

	allowed, explain, _ = e.EnforceEx("alice", "data2", "read")
	fmt.Printf("Request: alice, data2, read\n")
	fmt.Printf("Result: %v\n", allowed)
	fmt.Printf("Matched Policy: %v\n\n", explain)

	allowed, explain, _ = e.EnforceEx("bob", "data2", "write")
	fmt.Printf("Request: bob, data2, write\n")
	fmt.Printf("Result: %v\n", allowed)
	fmt.Printf("Matched Policy: %v\n", explain)
}

func testEnforce(e *casbin.Enforcer, sub, obj, act, description string) {
	allowed, err := e.Enforce(sub, obj, act)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	result := "❌ DENY"
	if allowed {
		result = "✓ ALLOW"
	}
	fmt.Printf("  %s can %s %s: %s\n", sub, act, obj, result)
	fmt.Printf("    → %s\n", description)
}
