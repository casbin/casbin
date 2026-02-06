// Copyright 2017 The casbin Authors. All Rights Reserved.
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

	"github.com/casbin/casbin/v3"
)

func main() {
	fmt.Println("=== RBAC with Resource Scope Demo ===\n")

	// Simple Resource Scope Example
	fmt.Println("--- Simple Resource Scope Example ---")
	e1, err := casbin.NewEnforcer("examples/rbac_with_resource_scope_model.conf", "examples/rbac_with_resource_scope_policy.csv")
	if err != nil {
		log.Fatalf("Failed to create enforcer: %v", err)
	}

	fmt.Println("\nInitial Policy:")
	fmt.Println("p, reader, resource1, read")
	fmt.Println("p, reader, resource2, read")
	fmt.Println("p, writer, resource1, write")
	fmt.Println("p, writer, resource2, write")
	fmt.Println("\nInitial Role Assignments:")
	fmt.Println("g, user1, reader, resource1")
	fmt.Println("g, user2, reader, resource2")
	fmt.Println("g, user3, writer, resource1")

	// Test enforcement
	fmt.Println("\n--- Enforcement Tests ---")
	testEnforce(e1, "user1", "resource1", "read", "user1 can read resource1")
	testEnforce(e1, "user1", "resource2", "read", "user1 can read resource2 (should be false)")
	testEnforce(e1, "user2", "resource1", "read", "user2 can read resource1 (should be false)")
	testEnforce(e1, "user2", "resource2", "read", "user2 can read resource2")
	testEnforce(e1, "user3", "resource1", "write", "user3 can write to resource1")
	testEnforce(e1, "user3", "resource2", "write", "user3 can write to resource2 (should be false)")

	// Get roles for users
	fmt.Println("\n--- Role Queries ---")
	printRoles(e1, "user1", "resource1")
	printRoles(e1, "user1", "resource2")
	printRoles(e1, "user2", "resource1")
	printRoles(e1, "user2", "resource2")

	// Add a new role assignment
	fmt.Println("\n--- Adding New Role Assignment ---")
	_, err = e1.AddRoleForUser("user4", "reader", "resource1")
	if err != nil {
		log.Fatalf("Failed to add role: %v", err)
	}
	fmt.Println("Added: g, user4, reader, resource1")
	testEnforce(e1, "user4", "resource1", "read", "user4 can read resource1")
	testEnforce(e1, "user4", "resource2", "read", "user4 can read resource2 (should be false)")

	// Multi-Tenant Resource Scope Example
	fmt.Println("\n\n--- Multi-Tenant Resource Scope Example ---")
	e2, err := casbin.NewEnforcer("examples/rbac_with_resource_scope_tenant_model.conf", "examples/rbac_with_resource_scope_tenant_policy.csv")
	if err != nil {
		log.Fatalf("Failed to create enforcer: %v", err)
	}

	fmt.Println("\nInitial Policy:")
	fmt.Println("p, reader, tenant1, resource1, read")
	fmt.Println("p, reader, tenant1, resource2, read")
	fmt.Println("p, reader, tenant2, resource1, read")
	fmt.Println("p, writer, tenant1, resource1, write")
	fmt.Println("\nInitial Role Assignments:")
	fmt.Println("g, user1, reader, tenant1::resource1")
	fmt.Println("g, user2, reader, tenant1::resource2")
	fmt.Println("g, user3, reader, tenant2::resource1")
	fmt.Println("g, user4, writer, tenant1::resource1")

	// Test enforcement with tenants
	fmt.Println("\n--- Enforcement Tests with Tenants ---")
	testEnforceWithTenant(e2, "user1", "tenant1", "resource1", "read", "user1 can read resource1 in tenant1")
	testEnforceWithTenant(e2, "user1", "tenant1", "resource2", "read", "user1 can read resource2 in tenant1 (should be false)")
	testEnforceWithTenant(e2, "user1", "tenant2", "resource1", "read", "user1 can read resource1 in tenant2 (should be false)")
	testEnforceWithTenant(e2, "user2", "tenant1", "resource2", "read", "user2 can read resource2 in tenant1")
	testEnforceWithTenant(e2, "user3", "tenant2", "resource1", "read", "user3 can read resource1 in tenant2")
	testEnforceWithTenant(e2, "user4", "tenant1", "resource1", "write", "user4 can write to resource1 in tenant1")

	// Get roles for users in tenant context
	fmt.Println("\n--- Role Queries with Tenants ---")
	printRoles(e2, "user1", "tenant1::resource1")
	printRoles(e2, "user1", "tenant1::resource2")
	printRoles(e2, "user2", "tenant1::resource2")
	printRoles(e2, "user3", "tenant2::resource1")

	// Demonstrate isolation
	fmt.Println("\n--- Demonstrating Multi-Tenancy Isolation ---")
	fmt.Println("Both user1 and user3 have 'reader' role, but for different tenant::resource combinations")
	testEnforceWithTenant(e2, "user1", "tenant1", "resource1", "read", "user1 can read resource1 in tenant1")
	testEnforceWithTenant(e2, "user1", "tenant2", "resource1", "read", "user1 can read resource1 in tenant2 (should be false - different tenant)")
	testEnforceWithTenant(e2, "user3", "tenant2", "resource1", "read", "user3 can read resource1 in tenant2")
	testEnforceWithTenant(e2, "user3", "tenant1", "resource1", "read", "user3 can read resource1 in tenant1 (should be false - different tenant)")

	fmt.Println("\n=== Demo Complete ===")
}

func testEnforce(e *casbin.Enforcer, sub, obj, act, description string) {
	result, err := e.Enforce(sub, obj, act)
	if err != nil {
		log.Printf("Error during enforcement: %v", err)
	}
	fmt.Printf("✓ %s: %t\n", description, result)
}

func testEnforceWithTenant(e *casbin.Enforcer, sub, tenant, obj, act, description string) {
	result, err := e.Enforce(sub, tenant, obj, act)
	if err != nil {
		log.Printf("Error during enforcement: %v", err)
	}
	fmt.Printf("✓ %s: %t\n", description, result)
}

func printRoles(e *casbin.Enforcer, user, scope string) {
	roles, err := e.GetRolesForUser(user, scope)
	if err != nil {
		log.Printf("Error getting roles: %v", err)
	}
	fmt.Printf("Roles for %s in scope '%s': %v\n", user, scope, roles)
}
