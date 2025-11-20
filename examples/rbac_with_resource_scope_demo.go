// Copyright 2025 The casbin Authors. All Rights Reserved.
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
	fmt.Println("=== RBAC with Resource Scope Demo ===\n")

	// Example 1: Simple Resource Scope
	fmt.Println("Example 1: Simple Resource Scope")
	fmt.Println("----------------------------------")
	
	e1, err := casbin.NewEnforcer("rbac_with_resource_scope_model.conf", "rbac_with_resource_scope_policy.csv")
	if err != nil {
		log.Fatal(err)
	}

	// user1 has reader role scoped to resource1
	ok, _ := e1.Enforce("user1", "resource1", "read")
	fmt.Printf("user1 can read resource1: %v\n", ok) // true

	ok, _ = e1.Enforce("user1", "resource2", "read")
	fmt.Printf("user1 can read resource2: %v\n", ok) // false

	// user2 has reader role scoped to resource2
	ok, _ = e1.Enforce("user2", "resource2", "read")
	fmt.Printf("user2 can read resource2: %v\n", ok) // true

	ok, _ = e1.Enforce("user2", "resource1", "read")
	fmt.Printf("user2 can read resource1: %v\n", ok) // false

	roles, _ := e1.GetRolesForUser("user1", "resource1")
	fmt.Printf("user1's roles for resource1: %v\n", roles)

	fmt.Println()

	// Example 2: Multi-Tenant with Resource Scope
	fmt.Println("Example 2: Multi-Tenant with Resource Scope")
	fmt.Println("--------------------------------------------")
	
	e2, err := casbin.NewEnforcer("rbac_with_resource_scope_multitenancy_model.conf", "rbac_with_resource_scope_multitenancy_policy.csv")
	if err != nil {
		log.Fatal(err)
	}

	// user1 has reader role for resource1 in tenant1
	ok, _ = e2.Enforce("user1", "tenant1", "resource1", "read")
	fmt.Printf("user1 can read resource1 in tenant1: %v\n", ok) // true

	ok, _ = e2.Enforce("user1", "tenant1", "resource2", "read")
	fmt.Printf("user1 can read resource2 in tenant1: %v\n", ok) // false

	ok, _ = e2.Enforce("user1", "tenant2", "resource1", "read")
	fmt.Printf("user1 can read resource1 in tenant2: %v\n", ok) // false

	// user3 has reader role for resource1 in tenant2
	ok, _ = e2.Enforce("user3", "tenant2", "resource1", "read")
	fmt.Printf("user3 can read resource1 in tenant2: %v\n", ok) // true

	roles, _ = e2.GetRolesForUser("user1", "tenant1::resource1")
	fmt.Printf("user1's roles for resource1 in tenant1: %v\n", roles)

	roles, _ = e2.GetRolesForUser("user3", "tenant2::resource1")
	fmt.Printf("user3's roles for resource1 in tenant2: %v\n", roles)

	fmt.Println()

	// Example 3: Dynamic Role Assignment
	fmt.Println("Example 3: Dynamic Role Assignment")
	fmt.Println("-----------------------------------")
	
	// Add a new role assignment with resource scope
	added, _ := e1.AddRoleForUser("user4", "writer", "resource1")
	fmt.Printf("Added writer role for user4 on resource1: %v\n", added)

	ok, _ = e1.Enforce("user4", "resource1", "write")
	fmt.Printf("user4 can write to resource1: %v\n", ok) // true

	ok, _ = e1.Enforce("user4", "resource2", "write")
	fmt.Printf("user4 can write to resource2: %v\n", ok) // false

	fmt.Println("\n=== Demo Complete ===")
}
