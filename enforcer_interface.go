// Copyright 2019 The casbin Authors. All Rights Reserved.
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
	"github.com/Knetic/govaluate"
	"github.com/casbin/casbin/v2/effect"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/casbin/casbin/v2/rbac"
)

var _ IEnforcer = &Enforcer{}
var _ IEnforcer = &SyncedEnforcer{}
var _ IEnforcer = &CachedEnforcer{}

// IEnforcer is the API interface of Enforcer
type IEnforcer interface {
	/* Enforcer API */
	InitWithFile(modelPath string, policyPath string) error
	InitWithAdapter(modelPath string, adapter persist.Adapter) error
	InitWithModelAndAdapter(m model.Model, adapter persist.Adapter) error
	initialize()
	LoadModel() error
	GetModel() model.Model
	SetModel(m model.Model)
	GetAdapter() persist.Adapter
	SetAdapter(adapter persist.Adapter)
	SetWatcher(watcher persist.Watcher) error
	GetRoleManager() rbac.RoleManager
	SetRoleManager(rm rbac.RoleManager)
	SetEffector(eft effect.Effector)
	ClearPolicy()
	LoadPolicy() error
	LoadFilteredPolicy(filter interface{}) error
	LoadIncrementalFilteredPolicy(filter interface{}) error
	IsFiltered() bool
	SavePolicy() error
	EnableEnforce(enable bool)
	EnableLog(enable bool)
	EnableAutoNotifyWatcher(enable bool)
	EnableAutoSave(autoSave bool)
	EnableAutoBuildRoleLinks(autoBuildRoleLinks bool)
	BuildRoleLinks() error
	enforce(matcher string, explains *[]string, rvals ...interface{}) (bool, error)
	Enforce(rvals ...interface{}) (bool, error)
	EnforceWithMatcher(matcher string, rvals ...interface{}) (bool, error)
	EnforceEx(rvals ...interface{}) (bool, []string, error)
	EnforceExWithMatcher(matcher string, rvals ...interface{}) (bool, []string, error)

	/* RBAC API */
	GetRolesForUser(name string, domain ...string) ([]string, error)
	GetUsersForRole(name string, domain ...string) ([]string, error)
	HasRoleForUser(name string, role string) (bool, error)
	AddRoleForUser(user string, role string) (bool, error)
	AddPermissionForUser(user string, permission ...string) (bool, error)
	DeletePermissionForUser(user string, permission ...string) (bool, error)
	DeletePermissionsForUser(user string) (bool, error)
	GetPermissionsForUser(user string) [][]string
	HasPermissionForUser(user string, permission ...string) bool
	GetImplicitRolesForUser(name string, domain ...string) ([]string, error)
	GetImplicitPermissionsForUser(user string, domain ...string) ([][]string, error)
	GetImplicitUsersForPermission(permission ...string) ([]string, error)
	DeleteRoleForUser(user string, role string) (bool, error)
	DeleteRolesForUser(user string) (bool, error)
	DeleteUser(user string) (bool, error)
	DeleteRole(role string) (bool, error)
	DeletePermission(permission ...string) (bool, error)

	/* RBAC API with domains*/
	GetUsersForRoleInDomain(name string, domain string) []string
	GetRolesForUserInDomain(name string, domain string) []string
	GetPermissionsForUserInDomain(user string, domain string) [][]string
	AddRoleForUserInDomain(user string, role string, domain string) (bool, error)
	DeleteRoleForUserInDomain(user string, role string, domain string) (bool, error)

	/* Management API */
	GetAllSubjects() []string
	GetAllNamedSubjects(ptype string) []string
	GetAllObjects() []string
	GetAllNamedObjects(ptype string) []string
	GetAllActions() []string
	GetAllNamedActions(ptype string) []string
	GetAllRoles() []string
	GetAllNamedRoles(ptype string) []string
	GetPolicy() [][]string
	GetFilteredPolicy(fieldIndex int, fieldValues ...string) [][]string
	GetNamedPolicy(ptype string) [][]string
	GetFilteredNamedPolicy(ptype string, fieldIndex int, fieldValues ...string) [][]string
	GetGroupingPolicy() [][]string
	GetFilteredGroupingPolicy(fieldIndex int, fieldValues ...string) [][]string
	GetNamedGroupingPolicy(ptype string) [][]string
	GetFilteredNamedGroupingPolicy(ptype string, fieldIndex int, fieldValues ...string) [][]string
	HasPolicy(params ...interface{}) bool
	HasNamedPolicy(ptype string, params ...interface{}) bool
	AddPolicy(params ...interface{}) (bool, error)
	AddPolicies(rules [][]string) (bool, error)
	AddNamedPolicy(ptype string, params ...interface{}) (bool, error)
	AddNamedPolicies(ptype string, rules [][]string) (bool, error)
	RemovePolicy(params ...interface{}) (bool, error)
	RemovePolicies(rules [][]string) (bool, error)
	RemoveFilteredPolicy(fieldIndex int, fieldValues ...string) (bool, error)
	RemoveNamedPolicy(ptype string, params ...interface{}) (bool, error)
	RemoveNamedPolicies(ptype string, rules [][]string) (bool, error)
	RemoveFilteredNamedPolicy(ptype string, fieldIndex int, fieldValues ...string) (bool, error)
	HasGroupingPolicy(params ...interface{}) bool
	HasNamedGroupingPolicy(ptype string, params ...interface{}) bool
	AddGroupingPolicy(params ...interface{}) (bool, error)
	AddGroupingPolicies(rules [][]string) (bool, error)
	AddNamedGroupingPolicy(ptype string, params ...interface{}) (bool, error)
	AddNamedGroupingPolicies(ptype string, rules [][]string) (bool, error)
	RemoveGroupingPolicy(params ...interface{}) (bool, error)
	RemoveGroupingPolicies(rules [][]string) (bool, error)
	RemoveFilteredGroupingPolicy(fieldIndex int, fieldValues ...string) (bool, error)
	RemoveNamedGroupingPolicy(ptype string, params ...interface{}) (bool, error)
	RemoveNamedGroupingPolicies(ptype string, rules [][]string) (bool, error)
	RemoveFilteredNamedGroupingPolicy(ptype string, fieldIndex int, fieldValues ...string) (bool, error)
	AddFunction(name string, function govaluate.ExpressionFunction)
}
