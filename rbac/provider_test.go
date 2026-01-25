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

package rbac

import (
	"testing"
)

// mockProvider is a mock implementation of Provider for testing purposes.
type mockProvider struct {
	users           []string
	roles           []string
	permissions     map[string][][]string
	rolePermissions map[string][][]string
}

func newMockProvider() *mockProvider {
	return &mockProvider{
		users:           []string{},
		roles:           []string{},
		permissions:     make(map[string][][]string),
		rolePermissions: make(map[string][][]string),
	}
}

// RoleManager interface methods - stub implementations for testing.
func (m *mockProvider) Clear() error { return nil }
func (m *mockProvider) AddLink(name1 string, name2 string, domain ...string) error {
	return nil
}
func (m *mockProvider) BuildRelationship(name1 string, name2 string, domain ...string) error {
	return nil
}
func (m *mockProvider) DeleteLink(name1 string, name2 string, domain ...string) error {
	return nil
}
func (m *mockProvider) HasLink(name1 string, name2 string, domain ...string) (bool, error) {
	return false, nil
}
func (m *mockProvider) GetRoles(name string, domain ...string) ([]string, error) {
	return nil, nil
}
func (m *mockProvider) GetUsers(name string, domain ...string) ([]string, error) {
	return nil, nil
}
func (m *mockProvider) GetImplicitRoles(name string, domain ...string) ([]string, error) {
	return nil, nil
}
func (m *mockProvider) GetImplicitUsers(name string, domain ...string) ([]string, error) {
	return nil, nil
}
func (m *mockProvider) GetDomains(name string) ([]string, error) { return nil, nil }
func (m *mockProvider) GetAllDomains() ([]string, error)         { return nil, nil }
func (m *mockProvider) PrintRoles() error                        { return nil }
func (m *mockProvider) Match(str string, pattern string) bool    { return false }
func (m *mockProvider) AddMatchingFunc(name string, fn MatchingFunc) {
}
func (m *mockProvider) AddDomainMatchingFunc(name string, fn MatchingFunc) {
}
func (m *mockProvider) DeleteDomain(domain string) error { return nil }

func (m *mockProvider) GetAllUsers() ([]string, error) {
	return m.users, nil
}

func (m *mockProvider) AddUser(user string) error {
	m.users = append(m.users, user)
	return nil
}

func (m *mockProvider) DeleteUser(user string) error {
	for i, u := range m.users {
		if u == user {
			m.users = append(m.users[:i], m.users[i+1:]...)
			break
		}
	}
	return nil
}

func (m *mockProvider) GetAllRoles() ([]string, error) {
	return m.roles, nil
}

func (m *mockProvider) AddRole(role string) error {
	m.roles = append(m.roles, role)
	return nil
}

func (m *mockProvider) DeleteRole(role string) error {
	for i, r := range m.roles {
		if r == role {
			m.roles = append(m.roles[:i], m.roles[i+1:]...)
			break
		}
	}
	return nil
}

func (m *mockProvider) GetPermissions(subject string) ([][]string, error) {
	return m.permissions[subject], nil
}

func (m *mockProvider) AddPermission(subject string, permission []string) error {
	m.permissions[subject] = append(m.permissions[subject], permission)
	return nil
}

func (m *mockProvider) DeletePermission(subject string, permission []string) error {
	// Simple implementation for testing
	return nil
}

func (m *mockProvider) GetRolePermissions(role string) ([][]string, error) {
	return m.rolePermissions[role], nil
}

func (m *mockProvider) AddRolePermission(role string, permission []string) error {
	m.rolePermissions[role] = append(m.rolePermissions[role], permission)
	return nil
}

func (m *mockProvider) DeleteRolePermission(role string, permission []string) error {
	// Simple implementation for testing
	return nil
}

// TestProviderInterface verifies that the Provider interface can be implemented and used.
func TestProviderInterface(t *testing.T) {
	var provider Provider = newMockProvider()

	// Test user operations
	err := provider.AddUser("alice")
	if err != nil {
		t.Fatalf("AddUser failed: %v", err)
	}

	err = provider.AddUser("bob")
	if err != nil {
		t.Fatalf("AddUser failed: %v", err)
	}

	users, err := provider.GetAllUsers()
	if err != nil {
		t.Fatalf("GetAllUsers failed: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("Expected 2 users, got %d", len(users))
	}

	err = provider.DeleteUser("alice")
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}

	users, err = provider.GetAllUsers()
	if err != nil {
		t.Fatalf("GetAllUsers failed: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("Expected 1 user after deletion, got %d", len(users))
	}

	// Test role operations
	err = provider.AddRole("admin")
	if err != nil {
		t.Fatalf("AddRole failed: %v", err)
	}

	err = provider.AddRole("editor")
	if err != nil {
		t.Fatalf("AddRole failed: %v", err)
	}

	roles, err := provider.GetAllRoles()
	if err != nil {
		t.Fatalf("GetAllRoles failed: %v", err)
	}
	if len(roles) != 2 {
		t.Fatalf("Expected 2 roles, got %d", len(roles))
	}

	err = provider.DeleteRole("admin")
	if err != nil {
		t.Fatalf("DeleteRole failed: %v", err)
	}

	roles, err = provider.GetAllRoles()
	if err != nil {
		t.Fatalf("GetAllRoles failed: %v", err)
	}
	if len(roles) != 1 {
		t.Fatalf("Expected 1 role after deletion, got %d", len(roles))
	}

	// Test permission operations
	err = provider.AddPermission("bob", []string{"data1", "read"})
	if err != nil {
		t.Fatalf("AddPermission failed: %v", err)
	}

	perms, err := provider.GetPermissions("bob")
	if err != nil {
		t.Fatalf("GetPermissions failed: %v", err)
	}
	if len(perms) != 1 {
		t.Fatalf("Expected 1 permission, got %d", len(perms))
	}

	// Test role permission operations
	err = provider.AddRolePermission("editor", []string{"data2", "write"})
	if err != nil {
		t.Fatalf("AddRolePermission failed: %v", err)
	}

	rolePerms, err := provider.GetRolePermissions("editor")
	if err != nil {
		t.Fatalf("GetRolePermissions failed: %v", err)
	}
	if len(rolePerms) != 1 {
		t.Fatalf("Expected 1 role permission, got %d", len(rolePerms))
	}
}
