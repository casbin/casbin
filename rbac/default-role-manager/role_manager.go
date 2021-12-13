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
	"github.com/casbin/casbin/v2/util"
)

const defaultDomain string = ""

type MatchingFunc func(arg1, arg2 string) bool

// RoleManager provides a default implementation for the RoleManager interface
type RoleManager struct {
	allDomains *sync.Map

	maxHierarchyLevel  int
	hasPattern         bool
	matchingFunc       MatchingFunc
	hasDomainPattern   bool
	domainMatchingFunc MatchingFunc

	logger log.Logger

	matchingFuncCache       *sync.Map
	domainMatchingFuncCache *sync.Map
}

// NewRoleManager is the constructor for creating an instance of the
// default RoleManager implementation.
func NewRoleManager(maxHierarchyLevel int) *RoleManager {
	rm := RoleManager{}
	rm.allDomains = &sync.Map{}
	rm.maxHierarchyLevel = maxHierarchyLevel
	rm.hasPattern = false
	rm.hasDomainPattern = false

	rm.SetLogger(&log.DefaultLogger{})

	return &rm
}

// AddMatchingFunc support use pattern in g
func (rm *RoleManager) AddMatchingFunc(name string, fn MatchingFunc) {
	rm.hasPattern = true
	rm.matchingFunc = fn
	rm.matchingFuncCache = &sync.Map{}
}

// AddDomainMatchingFunc support use domain pattern in g
func (rm *RoleManager) AddDomainMatchingFunc(name string, fn MatchingFunc) {
	rm.hasDomainPattern = true
	rm.domainMatchingFunc = fn
	rm.domainMatchingFuncCache = &sync.Map{}
}

// SetLogger sets role manager's logger.
func (rm *RoleManager) SetLogger(logger log.Logger) {
	rm.logger = logger
}

// Clear clears all stored data and resets the role manager to the initial state.
func (rm *RoleManager) Clear() error {
	rm.allDomains = &sync.Map{}
	rm.matchingFuncCache = &sync.Map{}
	rm.domainMatchingFuncCache = &sync.Map{}
	return nil
}

// AddLink adds the inheritance link between role: name1 and role: name2.
// aka role: name1 inherits role: name2.
func (rm *RoleManager) AddLink(name1 string, name2 string, domains ...string) error {
	switch len(domains) {
	case 0:
		domains = []string{defaultDomain}
		fallthrough
	case 1:
		domainValue, _ := rm.allDomains.LoadOrStore(domains[0], &Roles{})
		domain := domainValue.(*Roles)

		role1 := domain.createRole(name1)
		role2 := domain.createRole(name2)
		role1.addRole(role2)

		return nil
	default:
		return errors.ERR_DOMAIN_PARAMETER
	}

}

func (rm *RoleManager) BuildRelationship(name1, name2 string, domains ...string) error {
	switch len(domains) {
	case 0:
		domains = []string{defaultDomain}
		fallthrough
	case 1:
		domainValue, _ := rm.allDomains.LoadOrStore(domains[0], &Roles{})
		domain := domainValue.(*Roles)

		role1 := domain.createRole(name1)
		role2 := domain.createRole(name2)

		patternDomain := rm.getPatternDomain(domains[0])
		for _, domain := range patternDomain {
			allRoles, _ := rm.allDomains.LoadOrStore(domain, new(Roles))

			if rm.hasPattern {
				allRoles.(*Roles).Range(func(key, value interface{}) bool {
					namePattern := key.(string)
					if rm.match(namePattern, name1) {
						value.(*Role).addRole(role1)
					}
					if rm.match(name1, namePattern) {
						role1.addRole(value.(*Role))
					}
					if rm.match(namePattern, name2) {
						value.(*Role).addRole(role2)
					}
					if rm.match(name2, namePattern) {
						role2.addRole(value.(*Role))
					}
					return true
				})
			} else {
				if domain == domains[0] {
					continue
				}
				allRoles.(*Roles).Range(func(key, value interface{}) bool {
					if name1 == key.(string) {
						role1.addRole(value.(*Role))
					}
					if name2 == key.(string) {
						role2.addRole(value.(*Role))
					}
					return true
				})
			}
		}
		return nil
	default:
		return errors.ERR_DOMAIN_PARAMETER
	}
}

// DeleteLink deletes the inheritance link between role: name1 and role: name2.
// aka role: name1 does not inherit role: name2 any more.
func (rm *RoleManager) DeleteLink(name1 string, name2 string, domains ...string) error {
	switch len(domains) {
	case 0:
		domains = []string{defaultDomain}
		fallthrough
	case 1:

		domainValue, _ := rm.allDomains.LoadOrStore(domains[0], &Roles{})
		domain := domainValue.(*Roles)
		_, ok1 := domain.Load(name1)
		_, ok2 := domain.Load(name2)
		if !ok1 || !ok2 {
			return errors.ERR_NAMES12_NOT_FOUND
		}

		role1 := domain.createRole(name1)
		role2 := domain.createRole(name2)
		role1.deleteRole(role2)
		return nil
	default:
		return errors.ERR_DOMAIN_PARAMETER
	}
}

func (rm *RoleManager) getPatternDomain(domain string) []string {
	matchedDomains := []string{domain}
	if rm.hasDomainPattern {
		rm.allDomains.Range(func(key, value interface{}) bool {
			domainPattern := key.(string)

			if domain != domainPattern && rm.domainMatch(domain, domainPattern) {
				matchedDomains = append(matchedDomains, domainPattern)
			}
			return true
		})
	}
	return matchedDomains
}

// HasLink determines whether role: name1 inherits role: name2.
func (rm *RoleManager) HasLink(name1 string, name2 string, domains ...string) (bool, error) {
	switch len(domains) {
	case 0:
		domains = []string{defaultDomain}
		fallthrough
	case 1:
		if name1 == name2 {
			return true, nil
		}

		matchedDomain := rm.getPatternDomain(domains[0])

		roleQueue := []string{name1}
		inherited := make(map[string]bool)

		for len(roleQueue) != 0 {
			role := roleQueue[0]
			roleQueue = roleQueue[1:]
			if _, ok := inherited[role]; ok {
				continue
			}
			inherited[role] = true

			for _, domainName := range matchedDomain {

				domainValue, _ := rm.allDomains.LoadOrStore(domainName, &Roles{})
				domain := domainValue.(*Roles)

				if rm.hasPattern {
					flag := false
					domain.Range(func(key, value interface{}) bool {
						if rm.match(role, key.(string)) && value.(*Role).hasRoleWithMatchingFunc(name2, rm.maxHierarchyLevel, rm.match) {
							flag = true
							return false
						}
						return true
					})
					if flag {
						return true, nil
					}
				} else {
					role1Value, ok := domain.Load(role)
					if !ok {
						continue
					}
					role1 := role1Value.(*Role)
					result := role1.hasRole(name2, rm.maxHierarchyLevel)
					if result {
						return true, nil
					} else {
						for _, r := range role1.roles {
							roleQueue = append(roleQueue, r.name)
						}
					}
				}
			}
		}

		return false, nil
	default:
		return false, errors.ERR_DOMAIN_PARAMETER
	}
}

// GetRoles gets the roles that a subject inherits.
func (rm *RoleManager) GetRoles(name string, domains ...string) ([]string, error) {
	switch len(domains) {
	case 0:
		domains = []string{defaultDomain}
		fallthrough
	case 1:
		patternDomain := rm.getPatternDomain(domains[0])

		var gottenRoles []string
		for _, domain := range patternDomain {
			if !rm.hasRole(domain, name) {
				continue
			}
			domainValue, _ := rm.allDomains.LoadOrStore(domain, &Roles{})
			gottenRoles = append(gottenRoles, domainValue.(*Roles).createRole(name).getRoles()...)
		}
		gottenRoles = util.RemoveDuplicateElement(gottenRoles)
		return gottenRoles, nil
	default:
		return nil, errors.ERR_DOMAIN_PARAMETER
	}
}

// GetUsers gets the users that inherits a subject.
// domain is an unreferenced parameter here, may be used in other implementations.
func (rm *RoleManager) GetUsers(name string, domain ...string) ([]string, error) {
	switch len(domain) {
	case 0:
		domain = []string{defaultDomain}
		fallthrough
	case 1:
		patternDomain := rm.getPatternDomain(domain[0])

		var names []string
		for _, domain := range patternDomain {
			if !rm.hasRole(domain, name) {
				continue
			}

			domainValue, _ := rm.allDomains.LoadOrStore(domain, &Roles{})

			domainValue.(*Roles).Range(func(_, value interface{}) bool {
				role := value.(*Role)
				if role.hasDirectRole(name) {
					names = append(names, role.name)
				}
				return true
			})
		}

		return names, nil
	default:
		return nil, errors.ERR_DOMAIN_PARAMETER
	}
}

// PrintRoles prints all the roles to log.
func (rm *RoleManager) PrintRoles() error {
	if !(rm.logger).IsEnabled() {
		return nil
	}

	var roles []string
	rm.allDomains.Range(func(key, value interface{}) bool {
		value.(*Roles).Range(func(_, value interface{}) bool {
			if text := value.(*Role).toString(); text != "" {
				roles = append(roles, text)
			}
			return true
		})
		return true
	})
	
	rm.logger.LogRole(roles)
	return nil
}

// GetDomains gets domains that a user has
func (rm *RoleManager) GetDomains(name string) ([]string, error) {
	var domains []string
	rm.allDomains.Range(func(key, value interface{}) bool {
		domainName := key.(string)
		if rm.hasAnyRole(name, domainName) {
			domains = append(domains, domainName)
		}
		return true
	})
	return domains, nil
}

func (rm *RoleManager) hasAnyRole(name string, domain string) bool {
	patternDomain := rm.getPatternDomain(domain)
	for _, domain := range patternDomain {
		if rm.hasRole(domain, name) {
			return true
		}
	}
	return false
}

func (rm *RoleManager) hasRole(domain, name string) bool {
	roles, _ := rm.allDomains.LoadOrStore(domain, &Roles{})
	allRoles := roles.(*Roles)
	if rm.hasPattern {
		return allRoles.hasRole(domain, name, rm.match)
	} else {
		return allRoles.hasRole(domain, name, nil)
	}
}

func (rm *RoleManager) match(name1, name2 string) bool {
	cacheKey := strings.Join([]string{name1, name2}, "$$")
	if v, has := rm.matchingFuncCache.Load(cacheKey); has {
		return v.(bool)
	} else {
		matched := rm.matchingFunc(name1, name2)
		rm.matchingFuncCache.Store(cacheKey, matched)
		return matched
	}
}

func (rm *RoleManager) domainMatch(domain1, domain2 string) bool {
	cacheKey := strings.Join([]string{domain1, domain2}, "$$")
	if v, has := rm.domainMatchingFuncCache.Load(cacheKey); has {
		return v.(bool)
	} else {
		matched := rm.domainMatchingFunc(domain1, domain2)
		rm.domainMatchingFuncCache.Store(cacheKey, matched)
		return matched
	}
}

// Roles represents all roles in a domain
type Roles struct {
	sync.Map
}

func (roles *Roles) hasRole(domain, name string, matchingFunc MatchingFunc) bool {
	var ok bool
	if matchingFunc != nil {
		roles.Range(func(key, value interface{}) bool {
			namePattern := key.(string)
			if matchingFunc(name, namePattern) {
				ok = true
			}
			return true
		})
	} else {
		_, ok = roles.Load(name)
	}

	return ok
}

func (roles *Roles) createRole(name string) *Role {
	role, _ := roles.LoadOrStore(name, newRole(name))
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
	// determine whether this role has been added
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
	if r.hasDirectRole(name) {
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

func (r *Role) hasRoleWithMatchingFunc(name string, hierarchyLevel int, matchingFunc MatchingFunc) bool {
	if r.hasDirectRoleWithMatchingFunc(name, matchingFunc) {
		return true
	}

	if hierarchyLevel <= 0 {
		return false
	}

	for _, role := range r.roles {
		if role.hasRoleWithMatchingFunc(name, hierarchyLevel-1, matchingFunc) {
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

func (r *Role) hasDirectRoleWithMatchingFunc(name string, matchingFunc MatchingFunc) bool {
	for _, role := range r.roles {
		if role.name == name || (matchingFunc(name, role.name) && name != role.name) {
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
