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

package casbin

import (
	"errors"

	"github.com/Knetic/govaluate"
)

// DummyEnforcer is a dummy implementation of APIEnforcer which simply returns either
// nil or ErrorUnsupported on all function calls.
// DummyEnforcer is used to provide functionality to enforcer wrappers which do not
// implement APIEnforcer
type DummyEnforcer struct {
}

// ErrorUnsupported is returned to indicate that a specific function is not supported by an enforcer wrapper
var ErrorUnsupported = errors.New("unsupported function")

// GetRolesForUser returns nil and ErrorUnsupported
func (e *DummyEnforcer) GetRolesForUser(name string) ([]string, error) {
	return nil, ErrorUnsupported
}

// GetUsersForRole returns nil and ErrorUnsupported
func (e *DummyEnforcer) GetUsersForRole(name string) ([]string, error) {
	return nil, ErrorUnsupported
}

// HasRoleForUser returns false and ErrorUnsupported
func (e *DummyEnforcer) HasRoleForUser(name string, role string) (bool, error) {
	return false, ErrorUnsupported
}

// AddRoleForUser returns false and ErrorUnsupported
func (e *DummyEnforcer) AddRoleForUser(user string, role string) (bool, error) {
	return false, ErrorUnsupported
}

// DeleteRoleForUser returns false and ErrorUnsupported
func (e *DummyEnforcer) DeleteRoleForUser(user string, role string) (bool, error) {
	return false, ErrorUnsupported
}

// DeleteRolesForUser returns false and ErrorUnsupported
func (e *DummyEnforcer) DeleteRolesForUser(user string) (bool, error) {
	return false, ErrorUnsupported
}

// DeleteUser returns false and ErrorUnsupported
func (e *DummyEnforcer) DeleteUser(user string) (bool, error) {
	return false, ErrorUnsupported
}

// DeleteRole returns false and ErrorUnsupported
func (e *DummyEnforcer) DeleteRole(role string) (bool, error) {
	return false, ErrorUnsupported
}

// DeletePermission returns false and ErrorUnsupported
func (e *DummyEnforcer) DeletePermission(permission ...string) (bool, error) {
	return false, ErrorUnsupported
}

// AddPermissionForUser returns false and ErrorUnsupported
func (e *DummyEnforcer) AddPermissionForUser(user string, permission ...string) (bool, error) {
	return false, ErrorUnsupported
}

// DeletePermissionForUser returns false and ErrorUnsupported
func (e *DummyEnforcer) DeletePermissionForUser(user string, permission ...string) (bool, error) {
	return false, ErrorUnsupported
}

// DeletePermissionsForUser returns false and ErrorUnsupported
func (e *DummyEnforcer) DeletePermissionsForUser(user string) (bool, error) {
	return false, ErrorUnsupported
}

// GetPermissionsForUser returns nil
func (e *DummyEnforcer) GetPermissionsForUser(user string) [][]string {
	return nil
}

// HasPermissionForUser returns false
func (e *DummyEnforcer) HasPermissionForUser(user string, permission ...string) bool {
	return false
}

// GetImplicitRolesForUser returns nil and ErrorUnsupported
func (e *DummyEnforcer) GetImplicitRolesForUser(name string, domain ...string) ([]string, error) {
	return nil, ErrorUnsupported
}

// GetImplicitPermissionsForUser returns nil and ErrorUnsupported
func (e *DummyEnforcer) GetImplicitPermissionsForUser(user string, domain ...string) ([][]string, error) {
	return nil, ErrorUnsupported
}

// GetImplicitUsersForPermission returns nil and ErrorUnsupported
func (e *DummyEnforcer) GetImplicitUsersForPermission(permission ...string) ([]string, error) {
	return nil, ErrorUnsupported
}

// GetUsersForRoleInDomain returns nil
func (e *DummyEnforcer) GetUsersForRoleInDomain(name string, domain string) []string {
	return nil
}

// GetRolesForUserInDomain returns nil
func (e *DummyEnforcer) GetRolesForUserInDomain(name string, domain string) []string {
	return nil
}

// GetPermissionsForUserInDomain returns nil
func (e *DummyEnforcer) GetPermissionsForUserInDomain(user string, domain string) [][]string {
	return nil
}

// AddRoleForUserInDomain returns false and ErrorUnsupported
func (e *DummyEnforcer) AddRoleForUserInDomain(user string, role string, domain string) (bool, error) {
	return false, ErrorUnsupported
}

// DeleteRoleForUserInDomain returns false and ErrorUnsupported
func (e *DummyEnforcer) DeleteRoleForUserInDomain(user string, role string, domain string) (bool, error) {
	return false, ErrorUnsupported
}

// GetAllSubjects returns nil
func (e *DummyEnforcer) GetAllSubjects() []string {
	return nil
}

// GetAllNamedSubjects returns nil
func (e *DummyEnforcer) GetAllNamedSubjects(ptype string) []string {
	return nil
}

// GetAllObjects returns nil
func (e *DummyEnforcer) GetAllObjects() []string {
	return nil
}

// GetAllNamedObjects returns nil
func (e *DummyEnforcer) GetAllNamedObjects(ptype string) []string {
	return nil
}

// GetAllActions returns nil
func (e *DummyEnforcer) GetAllActions() []string {
	return nil
}

// GetAllNamedActions returns nil
func (e *DummyEnforcer) GetAllNamedActions(ptype string) []string {
	return nil
}

// GetAllRoles returns nil
func (e *DummyEnforcer) GetAllRoles() []string {
	return nil
}

// GetAllNamedRoles returns nil
func (e *DummyEnforcer) GetAllNamedRoles(ptype string) []string {
	return nil
}

// GetPolicy returns nil
func (e *DummyEnforcer) GetPolicy() [][]string {
	return nil
}

// GetFilteredPolicy returns nil
func (e *DummyEnforcer) GetFilteredPolicy(fieldIndex int, fieldValues ...string) [][]string {
	return nil
}

// GetNamedPolicy returns nil
func (e *DummyEnforcer) GetNamedPolicy(ptype string) [][]string {
	return nil
}

// GetFilteredNamedPolicy returns nil
func (e *DummyEnforcer) GetFilteredNamedPolicy(ptype string, fieldIndex int, fieldValues ...string) [][]string {
	return nil
}

// GetGroupingPolicy returns nil
func (e *DummyEnforcer) GetGroupingPolicy() [][]string {
	return nil
}

// GetFilteredGroupingPolicy returns nil
func (e *DummyEnforcer) GetFilteredGroupingPolicy(fieldIndex int, fieldValues ...string) [][]string {
	return nil
}

// GetNamedGroupingPolicy returns nil
func (e *DummyEnforcer) GetNamedGroupingPolicy(ptype string) [][]string {
	return nil
}

// GetFilteredNamedGroupingPolicy returns nil
func (e *DummyEnforcer) GetFilteredNamedGroupingPolicy(ptype string, fieldIndex int, fieldValues ...string) [][]string {
	return nil
}

// HasPolicy returns false
func (e *DummyEnforcer) HasPolicy(params ...interface{}) bool {
	return false
}

// HasNamedPolicy returns false
func (e *DummyEnforcer) HasNamedPolicy(ptype string, params ...interface{}) bool {
	return false
}

// AddPolicy returns false and ErrorUnsupported
func (e *DummyEnforcer) AddPolicy(params ...interface{}) (bool, error) {
	return false, ErrorUnsupported
}

// AddNamedPolicy returns false and ErrorUnsupported
func (e *DummyEnforcer) AddNamedPolicy(ptype string, params ...interface{}) (bool, error) {
	return false, ErrorUnsupported
}

// RemovePolicy returns false and ErrorUnsupported
func (e *DummyEnforcer) RemovePolicy(params ...interface{}) (bool, error) {
	return false, ErrorUnsupported
}

// RemoveFilteredPolicy returns false and ErrorUnsupported
func (e *DummyEnforcer) RemoveFilteredPolicy(fieldIndex int, fieldValues ...string) (bool, error) {
	return false, ErrorUnsupported
}

// RemoveNamedPolicy returns false and ErrorUnsupported
func (e *DummyEnforcer) RemoveNamedPolicy(ptype string, params ...interface{}) (bool, error) {
	return false, ErrorUnsupported
}

// RemoveFilteredNamedPolicy returns false and ErrorUnsupported
func (e *DummyEnforcer) RemoveFilteredNamedPolicy(ptype string, fieldIndex int, fieldValues ...string) (bool, error) {
	return false, ErrorUnsupported
}

// HasGroupingPolicy returns false
func (e *DummyEnforcer) HasGroupingPolicy(params ...interface{}) bool {
	return false
}

// HasNamedGroupingPolicy returns false
func (e *DummyEnforcer) HasNamedGroupingPolicy(ptype string, params ...interface{}) bool {
	return false
}

// AddGroupingPolicy returns false and ErrorUnsupported
func (e *DummyEnforcer) AddGroupingPolicy(params ...interface{}) (bool, error) {
	return false, ErrorUnsupported
}

// AddNamedGroupingPolicy returns false and ErrorUnsupported
func (e *DummyEnforcer) AddNamedGroupingPolicy(ptype string, params ...interface{}) (bool, error) {
	return false, ErrorUnsupported
}

// RemoveGroupingPolicy returns false and ErrorUnsupported
func (e *DummyEnforcer) RemoveGroupingPolicy(params ...interface{}) (bool, error) {
	return false, ErrorUnsupported
}

// RemoveFilteredGroupingPolicy returns false and ErrorUnsupported
func (e *DummyEnforcer) RemoveFilteredGroupingPolicy(fieldIndex int, fieldValues ...string) (bool, error) {
	return false, ErrorUnsupported
}

// RemoveNamedGroupingPolicy returns false and ErrorUnsupported
func (e *DummyEnforcer) RemoveNamedGroupingPolicy(ptype string, params ...interface{}) (bool, error) {
	return false, ErrorUnsupported
}

// RemoveFilteredNamedGroupingPolicy returns false and ErrorUnsupported
func (e *DummyEnforcer) RemoveFilteredNamedGroupingPolicy(ptype string, fieldIndex int, fieldValues ...string) (bool, error) {
	return false, ErrorUnsupported
}

// AddFunction does nothing
func (e *DummyEnforcer) AddFunction(name string, function govaluate.ExpressionFunction) {
	return
}
