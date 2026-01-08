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

package rbac

// Provider is the interface for Casbin providers.
// It provides identity info for Casbin, including users, roles, user-role-mappings,
// permissions, and role-permission-mappings.
// It extends the RoleManager interface because RoleManager only handles user-role-mappings,
// but Provider handles all identity and permission information.
// A Provider can be viewed as a way to import other auth permissions into Casbin.
// Provider implementations can be cloud providers (like AWS, Azure, GCP),
// identity vendors (like Okta, Auth0), or auth languages (like XACML).
type Provider interface {
	RoleManager

	// GetAllUsers gets all users.
	GetAllUsers() ([]string, error)
	// AddUser adds a user.
	AddUser(user string) error
	// DeleteUser deletes a user.
	DeleteUser(user string) error

	// GetAllRoles gets all roles.
	GetAllRoles() ([]string, error)
	// AddRole adds a role.
	AddRole(role string) error
	// DeleteRole deletes a role.
	DeleteRole(role string) error

	// GetPermissions gets the permissions for a subject (user or role).
	// Returns a list of permissions, where each permission is represented as a string slice
	// (e.g., [][]string{{"alice", "data1", "read"}, {"alice", "data2", "write"}}).
	GetPermissions(subject string) ([][]string, error)
	// AddPermission adds a permission for a subject (user or role).
	// The permission is represented as a string slice (e.g., []string{"alice", "data1", "read"}).
	AddPermission(subject string, permission []string) error
	// DeletePermission deletes a permission for a subject (user or role).
	DeletePermission(subject string, permission []string) error

	// GetRolePermissions gets the permissions for a role.
	// Returns a list of permissions, where each permission is represented as a string slice.
	GetRolePermissions(role string) ([][]string, error)
	// AddRolePermission adds a permission for a role.
	AddRolePermission(role string, permission []string) error
	// DeleteRolePermission deletes a permission for a role.
	DeleteRolePermission(role string, permission []string) error
}
