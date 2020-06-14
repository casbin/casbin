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

type MatchingFunc func(arg1, arg2 string) bool

// RoleManager provides a default implementation for the RoleManager interface
type RoleManager struct {
	allDomains         *sync.Map
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
	rm.allDomains = &sync.Map{}
	rm.allDomains.LoadOrStore(defaultDomain, new(Roles))
	rm.maxHierarchyLevel = maxHierarchyLevel
	rm.hasPattern = false
	rm.hasDomainPattern = false

	return &rm
}

// AddMatchingFunc support use pattern in g
func (rm *RoleManager) AddMatchingFunc(name string, fn MatchingFunc) {
	rm.hasPattern = true
	rm.matchingFunc = fn
}

// AddDomainMatchingFunc support use domain pattern in g
func (rm *RoleManager) AddDomainMatchingFunc(name string, fn MatchingFunc) {
	rm.hasDomainPattern = true
	rm.domainMatchingFunc = fn
}

func (rm *RoleManager) generateTempRoles(domain string) *Roles {
	rm.allDomains.LoadOrStore(domain, new(Roles))

	patternDomain := []string{domain}
	if rm.hasDomainPattern {
		rm.allDomains.Range(func(key, value interface{}) bool {
			if rm.domainMatchingFunc(domain, key.(string)) {
				patternDomain = append(patternDomain, key.(string))
			}
			return true
		})
	}

	allRoles := new(Roles)
	for _, domain := range patternDomain {
		value, _ := rm.allDomains.LoadOrStore(domain, new(Roles))
		value.(*Roles).Range(func(key, value interface{}) bool {
			role2 := value.(*Role)
			role1 := allRoles.createRole(role2.name, rm.matchingFunc)
			for _, v := range role2.roles {
				role3 := allRoles.createRole(v.name, rm.matchingFunc)
				role1.addRole(role3)
			}

			return true
		})
	}

	return allRoles
}

// Clear clears all stored data and resets the role manager to the initial state.
func (rm *RoleManager) Clear() error {
	rm.allDomains = &sync.Map{}
	rm.allDomains.LoadOrStore(defaultDomain, new(Roles))
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

	allRoles, _ := rm.allDomains.LoadOrStore(domain[0], new(Roles))

	role1, _ := allRoles.(*Roles).LoadOrStore(name1, newRole(name1))
	role2, _ := allRoles.(*Roles).LoadOrStore(name2, newRole(name2))
	role1.(*Role).addRole(role2.(*Role))
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

	allRoles, _ := rm.allDomains.LoadOrStore(domain[0], new(Roles))

	_, ok1 := allRoles.(*Roles).Load(name1)
	_, ok2 := allRoles.(*Roles).Load(name1)
	if !ok1 || !ok2 {
		return errors.ERR_NAMES12_NOT_FOUND
	}

	role1, _ := allRoles.(*Roles).LoadOrStore(name1, newRole(name1))
	role2, _ := allRoles.(*Roles).LoadOrStore(name2, newRole(name2))
	role1.(*Role).deleteRole(role2.(*Role))
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

	var allRoles *Roles
	if rm.hasDomainPattern || rm.hasPattern {
		allRoles = rm.generateTempRoles(domain[0])
	} else {
		roles, _ := rm.allDomains.LoadOrStore(domain[0], new(Roles))
		allRoles = roles.(*Roles)
	}

	if !allRoles.hasRole(name1, rm.matchingFunc) || !allRoles.hasRole(name2, rm.matchingFunc) {
		return false, nil
	}

	role1 := allRoles.createRole(name1, rm.matchingFunc)
	return role1.hasRole(name2, rm.maxHierarchyLevel), nil
}

// GetRoles gets the roles that a subject inherits.
func (rm *RoleManager) GetRoles(name string, domain ...string) ([]string, error) {
	if len(domain) == 0 {
		domain = []string{defaultDomain}
	} else if len(domain) > 1 {
		return nil, errors.ERR_DOMAIN_PARAMETER
	}

	var allRoles *Roles
	if rm.hasDomainPattern || rm.hasPattern {
		allRoles = rm.generateTempRoles(domain[0])
	} else {
		roles, _ := rm.allDomains.LoadOrStore(domain[0], new(Roles))
		allRoles = roles.(*Roles)
	}

	if !allRoles.hasRole(name, rm.matchingFunc) {
		return []string{}, nil
	}

	roles := allRoles.createRole(name, rm.matchingFunc).getRoles()

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

	var allRoles *Roles
	if rm.hasDomainPattern || rm.hasPattern {
		allRoles = rm.generateTempRoles(domain[0])
	} else {
		roles, _ := rm.allDomains.LoadOrStore(domain[0], new(Roles))
		allRoles = roles.(*Roles)
	}

	if !allRoles.hasRole(name, rm.domainMatchingFunc) {
		return nil, errors.ERR_NAME_NOT_FOUND
	}

	names := []string{}
	allRoles.Range(func(_, value interface{}) bool {
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
		rm.allDomains.Range(func(_, value interface{}) bool {
			value.(*Roles).Range(func(_, value interface{}) bool {
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

			return true
		})
		log.LogPrint(sb.String())
	}
	return nil
}

// Roles represents all roles in a domain
type Roles struct {
	sync.Map
}

func (roles *Roles) hasRole(name string, matchingFunc MatchingFunc) bool {
	var ok bool
	if matchingFunc != nil {
		roles.Range(func(key, value interface{}) bool {
			if matchingFunc(name, key.(string)) {
				ok = true
			}
			return true
		})
	} else {
		_, ok = roles.Load(name)
	}

	return ok
}

func (roles *Roles) createRole(name string, matchingFunc MatchingFunc) *Role {
	role, _ := roles.LoadOrStore(name, newRole(name))

	if matchingFunc != nil {
		roles.Range(func(key, value interface{}) bool {
			if matchingFunc(name, key.(string)) && name != key.(string) {
				// Add new role to matching role
				role1, _ := roles.LoadOrStore(key.(string), newRole(key.(string)))
				role.(*Role).addRole(role1.(*Role))
			}
			return true
		})
	}

	return role.(*Role)
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
