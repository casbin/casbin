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

package defaultrolemanager

import (
	"strings"
	"sync"

	"github.com/casbin/casbin/v2/errors"
	"github.com/casbin/casbin/v2/log"
	"github.com/casbin/casbin/v2/rbac"
)

const defaultDomain string = "casbin::default"
const tempDomain string = "casbin::temp"

type MatchingFunc func(arg1, arg2 string) bool

// RoleManager provides a default implementation for the RoleManager interface
type RoleManager struct {
	allRoles           map[string]*sync.Map
	maxHierarchyLevel  int
	hasPattern         bool
	matchingFunc       MatchingFunc
	hasDomainPattern   bool
	domainMatchingFunc MatchingFunc
}

// NewRoleManager is the constructor for creating an instance of the
// default RoleManager implementation.
func NewRoleManager(maxHierarchyLevel int) rbac.RoleManager {
	rm := RoleManager{}
	rm.allRoles = make(map[string]*sync.Map)
	rm.allRoles[defaultDomain] = &sync.Map{}
	rm.maxHierarchyLevel = maxHierarchyLevel
	rm.hasPattern = false
	rm.hasDomainPattern = false

	return &rm
}

// e.BuildRoleLinks must be called after AddMatchingFunc().
//
// example: e.GetRoleManager().(*defaultrolemanager.RoleManager).AddMatchingFunc('matcher', util.KeyMatch)
func (rm *RoleManager) AddMatchingFunc(name string, fn MatchingFunc, forDomain ...bool) error {
	if len(forDomain) == 1 && forDomain[0] {
		rm.hasDomainPattern = true
		rm.domainMatchingFunc = fn
		return nil
	} else if len(forDomain) > 1 {
		return errors.ERR_USE_DOMAIN_PARAMETER
	}

	rm.hasPattern = true
	rm.matchingFunc = fn
	return nil
}

func (rm *RoleManager) generateTempDomain(domain string) {
	if _, ok := rm.allRoles[domain]; !ok {
		rm.allRoles[domain] = &sync.Map{}
	}

	patternDomain := []string{domain}
	for k := range rm.allRoles {
		if rm.domainMatchingFunc(domain, k) {
			patternDomain = append(patternDomain, k)
		}
	}

	for _, domain := range patternDomain {
		rm.allRoles[domain].Range(func(key, value interface{}) bool {
			role2 := value.(*Role)
			for _, v := range role2.roles {
				rm.AddLink(role2.name, v.name, tempDomain)
			}

			return true
		})
	}
}

func (rm *RoleManager) hasRole(name string, domain string) bool {
	var ok bool
	if rm.hasPattern {
		rm.allRoles[domain].Range(func(key, value interface{}) bool {
			if rm.matchingFunc(name, key.(string)) {
				ok = true
			}
			return true
		})
	} else {
		_, ok = rm.allRoles[domain].Load(name)
	}

	return ok
}

func (rm *RoleManager) createRole(name string, domain string) *Role {
	if _, ok := rm.allRoles[domain]; !ok {
		rm.allRoles[domain] = &sync.Map{}
	}

	role, _ := rm.allRoles[domain].LoadOrStore(name, newRole(name))

	if rm.hasPattern {
		rm.allRoles[domain].Range(func(key, value interface{}) bool {
			if rm.matchingFunc(name, key.(string)) && name != key.(string) {
				// Add new role to matching role
				role1, _ := rm.allRoles[domain].LoadOrStore(key.(string), newRole(key.(string)))
				role.(*Role).addRole(role1.(*Role))
			}
			return true
		})
	}

	return role.(*Role)
}

// Clear clears all stored data and resets the role manager to the initial state.
func (rm *RoleManager) Clear() error {
	rm.allRoles = make(map[string]*sync.Map)
	rm.allRoles[defaultDomain] = &sync.Map{}
	return nil
}

// AddLink adds the inheritance link between role: name1 and role: name2.
// aka role: name1 inherits role: name2.
func (rm *RoleManager) AddLink(name1 string, name2 string, domain ...string) error {
	if len(domain) == 0 {
		domain = []string{defaultDomain}
	} else if len(domain) > 1 {
		return errors.ERR_DOMAIN_PARAMETER
	}

	role1 := rm.createRole(name1, domain[0])
	role2 := rm.createRole(name2, domain[0])
	role1.addRole(role2)
	return nil
}

// DeleteLink deletes the inheritance link between role: name1 and role: name2.
// aka role: name1 does not inherit role: name2 any more.
func (rm *RoleManager) DeleteLink(name1 string, name2 string, domain ...string) error {
	if len(domain) == 0 {
		domain = []string{defaultDomain}
	} else if len(domain) > 1 {
		return errors.ERR_DOMAIN_PARAMETER
	}

	if !rm.hasRole(name1, domain[0]) || !rm.hasRole(name2, domain[0]) {
		return errors.ERR_NAMES12_NOT_FOUND
	}

	role1 := rm.createRole(name1, domain[0])
	role2 := rm.createRole(name2, domain[0])
	role1.deleteRole(role2)
	return nil
}

// HasLink determines whether role: name1 inherits role: name2.
func (rm *RoleManager) HasLink(name1 string, name2 string, domain ...string) (bool, error) {
	if len(domain) == 0 {
		domain = []string{defaultDomain}
	} else if len(domain) > 1 {
		return false, errors.ERR_DOMAIN_PARAMETER
	}

	if name1 == name2 {
		return true, nil
	}

	if rm.hasDomainPattern {
		rm.generateTempDomain(domain[0])
		domain = []string{tempDomain}
		defer func() {
			rm.allRoles[tempDomain] = &sync.Map{}
		}()
	}

	if !rm.hasRole(name1, domain[0]) || !rm.hasRole(name2, domain[0]) {
		return false, nil
	}

	role1 := rm.createRole(name1, domain[0])
	return role1.hasRole(name2, rm.maxHierarchyLevel), nil
}

// GetRoles gets the roles that a subject inherits.
func (rm *RoleManager) GetRoles(name string, domain ...string) ([]string, error) {
	if len(domain) == 0 {
		domain = []string{defaultDomain}
	} else if len(domain) > 1 {
		return nil, errors.ERR_DOMAIN_PARAMETER
	}

	if rm.hasDomainPattern {
		rm.generateTempDomain(domain[0])
		domain = []string{tempDomain}
		defer func() {
			rm.allRoles[tempDomain] = &sync.Map{}
		}()
	}

	if !rm.hasRole(name, domain[0]) {
		return []string{}, nil
	}

	roles := rm.createRole(name, domain[0]).getRoles()

	return roles, nil
}

// GetUsers gets the users that inherits a subject.
// domain is an unreferenced parameter here, may be used in other implementations.
func (rm *RoleManager) GetUsers(name string, domain ...string) ([]string, error) {
	if len(domain) == 0 {
		domain = []string{defaultDomain}
	} else if len(domain) > 1 {
		return nil, errors.ERR_DOMAIN_PARAMETER
	}

	if rm.hasDomainPattern {
		rm.generateTempDomain(domain[0])
		domain = []string{tempDomain}
		defer func() {
			rm.allRoles[tempDomain] = &sync.Map{}
		}()
	}

	if !rm.hasRole(name, domain[0]) {
		return nil, errors.ERR_NAME_NOT_FOUND
	}

	names := []string{}
	rm.allRoles[domain[0]].Range(func(_, value interface{}) bool {
		role := value.(*Role)
		if role.hasDirectRole(name) {
			names = append(names, role.name)
		}
		return true
	})

	return names, nil
}

// PrintRoles prints all the roles to log.
func (rm *RoleManager) PrintRoles() error {
	if log.GetLogger().IsEnabled() {
		var sb strings.Builder
		for _, value := range rm.allRoles {
			value.Range(func(_, value interface{}) bool {
				if text := value.(*Role).toString(); text != "" {
					if sb.Len() == 0 {
						sb.WriteString(text)
					} else {
						sb.WriteString(", ")
						sb.WriteString(text)
					}
				}
				return true
			})
		}
		log.LogPrint(sb.String())
	}
	return nil
}

// Role represents the data structure for a role in RBAC.
type Role struct {
	name  string
	roles []*Role
}

func newRole(name string) *Role {
	r := Role{}
	r.name = name
	return &r
}

func (r *Role) addRole(role *Role) {
	for _, rr := range r.roles {
		if rr.name == role.name {
			return
		}
	}

	r.roles = append(r.roles, role)
}

func (r *Role) deleteRole(role *Role) {
	for i, rr := range r.roles {
		if rr.name == role.name {
			r.roles = append(r.roles[:i], r.roles[i+1:]...)
			return
		}
	}
}

func (r *Role) hasRole(name string, hierarchyLevel int) bool {
	if r.name == name {
		return true
	}

	if hierarchyLevel <= 0 {
		return false
	}

	for _, role := range r.roles {
		if role.hasRole(name, hierarchyLevel-1) {
			return true
		}
	}
	return false
}

func (r *Role) hasDirectRole(name string) bool {
	for _, role := range r.roles {
		if role.name == name {
			return true
		}
	}

	return false
}

func (r *Role) toString() string {
	if len(r.roles) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(r.name)
	sb.WriteString(" < ")
	if len(r.roles) != 1 {
		sb.WriteString("(")
	}

	for i, role := range r.roles {
		if i == 0 {
			sb.WriteString(role.name)
		} else {
			sb.WriteString(", ")
			sb.WriteString(role.name)
		}
	}

	if len(r.roles) != 1 {
		sb.WriteString(")")
	}

	return sb.String()
}

func (r *Role) getRoles() []string {
	names := []string{}
	for _, role := range r.roles {
		names = append(names, role.name)
	}
	return names
}
