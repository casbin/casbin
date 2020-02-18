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

type DummyEnforcer struct {
}

var ErrorUnsupported = errors.New("unsupported function")

func (e *DummyEnforcer) GetRolesForUser(name string) ([]string, error) {
	return nil, ErrorUnsupported
}

func (e *DummyEnforcer) GetUsersForRole(name string) ([]string, error) {
	return nil, ErrorUnsupported
}

func (e *DummyEnforcer) HasRoleForUser(name string, role string) (bool, error) {
	return false, ErrorUnsupported
}

func (e *DummyEnforcer) AddRoleForUser(user string, role string) (bool, error) {
	return false, ErrorUnsupported
}

func (e *DummyEnforcer) DeleteRoleForUser(user string, role string) (bool, error) {
	return false, ErrorUnsupported
}

func (e *DummyEnforcer) DeleteRolesForUser(user string) (bool, error) {
	return false, ErrorUnsupported
}

func (e *DummyEnforcer) DeleteUser(user string) (bool, error) {
	return false, ErrorUnsupported
}

func (e *DummyEnforcer) DeleteRole(role string) (bool, error) {
	return false, ErrorUnsupported
}

func (e *DummyEnforcer) DeletePermission(permission ...string) (bool, error) {
	return false, ErrorUnsupported
}

func (e *DummyEnforcer) AddPermissionForUser(user string, permission ...string) (bool, error) {
	return false, ErrorUnsupported
}

func (e *DummyEnforcer) DeletePermissionForUser(user string, permission ...string) (bool, error) {
	return false, ErrorUnsupported
}

func (e *DummyEnforcer) DeletePermissionsForUser(user string) (bool, error) {
	return false, ErrorUnsupported
}

func (e *DummyEnforcer) GetPermissionsForUser(user string) [][]string {
	return nil
}

func (e *DummyEnforcer) HasPermissionForUser(user string, permission ...string) bool {
	return false
}

func (e *DummyEnforcer) GetImplicitRolesForUser(name string, domain ...string) ([]string, error) {
	return nil, ErrorUnsupported
}

func (e *DummyEnforcer) GetImplicitPermissionsForUser(user string, domain ...string) ([][]string, error) {
	return nil, ErrorUnsupported
}

func (e *DummyEnforcer) GetImplicitUsersForPermission(permission ...string) ([]string, error) {
	return nil, ErrorUnsupported
}

func (e *DummyEnforcer) GetUsersForRoleInDomain(name string, domain string) []string {
	return nil
}

func (e *DummyEnforcer) GetRolesForUserInDomain(name string, domain string) []string {
	return nil
}

func (e *DummyEnforcer) GetPermissionsForUserInDomain(user string, domain string) [][]string {
	return nil
}

func (e *DummyEnforcer) AddRoleForUserInDomain(user string, role string, domain string) (bool, error) {
	return false, ErrorUnsupported
}

func (e *DummyEnforcer) DeleteRoleForUserInDomain(user string, role string, domain string) (bool, error) {
	return false, ErrorUnsupported
}

func (e *DummyEnforcer) GetAllSubjects() []string {
	return nil
}

func (e *DummyEnforcer) GetAllNamedSubjects(ptype string) []string {
	return nil
}

func (e *DummyEnforcer) GetAllObjects() []string {
	return nil
}

func (e *DummyEnforcer) GetAllNamedObjects(ptype string) []string {
	return nil
}

func (e *DummyEnforcer) GetAllActions() []string {
	return nil
}

func (e *DummyEnforcer) GetAllNamedActions(ptype string) []string {
	return nil
}

func (e *DummyEnforcer) GetAllRoles() []string {
	return nil
}

func (e *DummyEnforcer) GetAllNamedRoles(ptype string) []string {
	return nil
}

func (e *DummyEnforcer) GetPolicy() [][]string {
	return nil
}

func (e *DummyEnforcer) GetFilteredPolicy(fieldIndex int, fieldValues ...string) [][]string {
	return nil
}

func (e *DummyEnforcer) GetNamedPolicy(ptype string) [][]string {
	return nil
}

func (e *DummyEnforcer) GetFilteredNamedPolicy(ptype string, fieldIndex int, fieldValues ...string) [][]string {
	return nil
}

func (e *DummyEnforcer) GetGroupingPolicy() [][]string {
	return nil
}

func (e *DummyEnforcer) GetFilteredGroupingPolicy(fieldIndex int, fieldValues ...string) [][]string {
	return nil
}

func (e *DummyEnforcer) GetNamedGroupingPolicy(ptype string) [][]string {
	return nil
}

func (e *DummyEnforcer) GetFilteredNamedGroupingPolicy(ptype string, fieldIndex int, fieldValues ...string) [][]string {
	return nil
}

func (e *DummyEnforcer) HasPolicy(params ...interface{}) bool {
	return false
}

func (e *DummyEnforcer) HasNamedPolicy(ptype string, params ...interface{}) bool {
	return false
}

func (e *DummyEnforcer) AddPolicy(params ...interface{}) (bool, error) {
	return false, ErrorUnsupported
}

func (e *DummyEnforcer) AddNamedPolicy(ptype string, params ...interface{}) (bool, error) {
	return false, ErrorUnsupported
}

func (e *DummyEnforcer) RemovePolicy(params ...interface{}) (bool, error) {
	return false, ErrorUnsupported
}

func (e *DummyEnforcer) RemoveFilteredPolicy(fieldIndex int, fieldValues ...string) (bool, error) {
	return false, ErrorUnsupported
}

func (e *DummyEnforcer) RemoveNamedPolicy(ptype string, params ...interface{}) (bool, error) {
	return false, ErrorUnsupported
}

func (e *DummyEnforcer) RemoveFilteredNamedPolicy(ptype string, fieldIndex int, fieldValues ...string) (bool, error) {
	return false, ErrorUnsupported
}

func (e *DummyEnforcer) HasGroupingPolicy(params ...interface{}) bool {
	return false
}

func (e *DummyEnforcer) HasNamedGroupingPolicy(ptype string, params ...interface{}) bool {
	return false
}

func (e *DummyEnforcer) AddGroupingPolicy(params ...interface{}) (bool, error) {
	return false, ErrorUnsupported
}

func (e *DummyEnforcer) AddNamedGroupingPolicy(ptype string, params ...interface{}) (bool, error) {
	return false, ErrorUnsupported
}

func (e *DummyEnforcer) RemoveGroupingPolicy(params ...interface{}) (bool, error) {
	return false, ErrorUnsupported
}

func (e *DummyEnforcer) RemoveFilteredGroupingPolicy(fieldIndex int, fieldValues ...string) (bool, error) {
	return false, ErrorUnsupported
}

func (e *DummyEnforcer) RemoveNamedGroupingPolicy(ptype string, params ...interface{}) (bool, error) {
	return false, ErrorUnsupported
}

func (e *DummyEnforcer) RemoveFilteredNamedGroupingPolicy(ptype string, fieldIndex int, fieldValues ...string) (bool, error) {
	return false, ErrorUnsupported
}

func (e *DummyEnforcer) AddFunction(name string, function govaluate.ExpressionFunction) {
	return
}
