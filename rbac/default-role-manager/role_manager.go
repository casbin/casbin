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
	"fmt"
	"strings"
	"sync"

	"github.com/casbin/casbin/v2/errors"
	"github.com/casbin/casbin/v2/log"
	"github.com/casbin/casbin/v2/util"
)

const defaultDomain string = ""
const defaultSeparator = "::"

type MatchingFunc func(arg1, arg2 string) bool

// RoleManager provides a default implementation for the RoleManager interface
type RoleManager struct {
	roles              *Roles
	domains            map[string]struct{}
	maxHierarchyLevel  int
	hasPattern         bool
	matchingFunc       MatchingFunc
	hasDomainPattern   bool
	domainMatchingFunc MatchingFunc

	logger log.Logger
}

// NewRoleManager is the constructor for creating an instance of the
// default RoleManager implementation.
func NewRoleManager(maxHierarchyLevel int) *RoleManager {
	rm := RoleManager{}
	rm.roles = &Roles{sync.Map{}}
	rm.domains = make(map[string]struct{})
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
}

// AddDomainMatchingFunc support use domain pattern in g
func (rm *RoleManager) AddDomainMatchingFunc(name string, fn MatchingFunc) {
	rm.hasDomainPattern = true
	rm.domainMatchingFunc = fn
}

// SetLogger sets role manager's logger.
func (rm *RoleManager) SetLogger(logger log.Logger) {
	rm.logger = logger
}

// Clear clears all stored data and resets the role manager to the initial state.
func (rm *RoleManager) Clear() error {
	rm.roles = &Roles{sync.Map{}}
	rm.domains = make(map[string]struct{})
	return nil
}

// AddLink adds the inheritance link between role: name1 and role: name2.
// aka role: name1 inherits role: name2.
func (rm *RoleManager) AddLink(name1 string, name2 string, domain ...string) error {
	switch len(domain) {
	case 0:
		domain = []string{defaultDomain}
		fallthrough
	case 1:
		rm.domains[domain[0]] = struct{}{}
		patternDomain := rm.getPatternDomain(domain[0])

		for _, domain := range patternDomain {
			name1WithDomain := getNameWithDomain(domain, name1)
			name2WithDomain := getNameWithDomain(domain, name2)

			role1 := rm.roles.createRole(name1WithDomain)
			role2 := rm.roles.createRole(name2WithDomain)
			role1.addRole(role2)

			if rm.hasPattern {
				rm.roles.Range(func(key, value interface{}) bool {
					domainPattern, namePattern := getNameAndDomain(key.(string))
					if rm.hasDomainPattern {
						if !rm.domainMatchingFunc(domainPattern, domain) {
							return true
						}
					} else {
						if domainPattern != domain {
							return true
						}
					}
					if rm.matchingFunc(namePattern, name1) && name1 != namePattern && name2 != namePattern {
						valueRole, _ := rm.roles.LoadOrStore(key.(string), newRole(key.(string)))
						valueRole.(*Role).addRole(role1)
					}
					if rm.matchingFunc(namePattern, name2) && name2 != namePattern && name1 != namePattern {
						role2.addRole(value.(*Role))
					}
					if rm.matchingFunc(name1, namePattern) && name1 != namePattern && name2 != namePattern {
						valueRole, _ := rm.roles.LoadOrStore(key.(string), newRole(key.(string)))
						valueRole.(*Role).addRole(role1)
					}
					if rm.matchingFunc(name2, namePattern) && name2 != namePattern && name1 != namePattern {
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
func (rm *RoleManager) DeleteLink(name1 string, name2 string, domain ...string) error {
	switch len(domain) {
	case 0:
		domain = []string{defaultDomain}
		fallthrough
	case 1:
		name1WithDomain := getNameWithDomain(domain[0], name1)
		name2WithDomain := getNameWithDomain(domain[0], name2)

		_, ok1 := rm.roles.Load(name1WithDomain)
		_, ok2 := rm.roles.Load(name2WithDomain)
		if !ok1 || !ok2 {
			return errors.ERR_NAMES12_NOT_FOUND
		}

		role1 := rm.roles.createRole(name1WithDomain)
		role2 := rm.roles.createRole(name2WithDomain)
		role1.deleteRole(role2)
		return nil
	default:
		return errors.ERR_DOMAIN_PARAMETER
	}
}

func (rm *RoleManager) getPatternDomain(domain string) []string {
	matchedDomains := []string{domain}
	if rm.hasDomainPattern {
		for domainPattern := range rm.domains {
			if domain != domainPattern && rm.domainMatchingFunc(domain, domainPattern) {
				matchedDomains = append(matchedDomains, domainPattern)
			}
		}
	}
	return matchedDomains
}

// HasLink determines whether role: name1 inherits role: name2.
func (rm *RoleManager) HasLink(name1 string, name2 string, domain ...string) (bool, error) {
	switch len(domain) {
	case 0:
		domain = []string{defaultDomain}
		fallthrough
	case 1:
		if name1 == name2 {
			return true, nil
		}

		matchedDomain := rm.getPatternDomain(domain[0])

		for _, domain := range matchedDomain {
			if !rm.roles.hasRole(domain, name1, rm.matchingFunc) || !rm.roles.hasRole(domain, name2, rm.matchingFunc) {
				continue
			}

			name1WithDomain := getNameWithDomain(domain, name1)
			name2WithDomain := getNameWithDomain(domain, name2)

			if rm.hasPattern {
				flag := false
				rm.roles.Range(func(key, value interface{}) bool {
					nameWithDomain := key.(string)
					keyDomain, name := getNameAndDomain(nameWithDomain)
					if rm.hasDomainPattern {
						if !rm.domainMatchingFunc(domain, keyDomain) {
							return true
						}
					} else if domain != keyDomain {
						return true
					}
					if rm.matchingFunc(name1, name) && value.(*Role).hasRoleWithMatchingFunc(name2, rm.maxHierarchyLevel, rm.matchingFunc) {
						flag = true
						return false
					}
					return true
				})
				if flag {
					return true, nil
				}
			} else {
				role1 := rm.roles.createRole(name1WithDomain)
				result := role1.hasRole(name2WithDomain, rm.maxHierarchyLevel)
				if result {
					return true, nil
				}
			}
		}
		return false, nil
	default:
		return false, errors.ERR_DOMAIN_PARAMETER
	}
}

// GetRoles gets the roles that a subject inherits.
func (rm *RoleManager) GetRoles(name string, domain ...string) ([]string, error) {
	switch len(domain) {
	case 0:
		domain = []string{defaultDomain}
		fallthrough
	case 1:
		patternDomain := rm.getPatternDomain(domain[0])

		var gottenRoles []string
		for _, domain := range patternDomain {
			nameWithDomain := getNameWithDomain(domain, name)
			if !rm.roles.hasRole(domain, name, rm.matchingFunc) {
				continue
			}
			gottenRoles = append(gottenRoles, rm.roles.createRole(nameWithDomain).getRoles()...)
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
			nameWithDomain := getNameWithDomain(domain, name)
			if !rm.roles.hasRole(domain, name, rm.matchingFunc) {
				return nil, errors.ERR_NAME_NOT_FOUND
			}

			rm.roles.Range(func(_, value interface{}) bool {
				role := value.(*Role)
				if role.hasDirectRole(nameWithDomain) {
					_, roleName := getNameAndDomain(role.nameWithDomain)
					names = append(names, roleName)
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
	rm.roles.Range(func(_, value interface{}) bool {
		if text := value.(*Role).toString(); text != "" {
			roles = append(roles, text)
		}
		return true
	})

	rm.logger.LogRole(roles)
	return nil
}

// GetDomains gets domains that a user has
func (rm *RoleManager) GetDomains(name string) ([]string, error) {
	var domains []string
	for domain := range rm.domains {
		if rm.hasAnyRole(name, domain) {
			domains = append(domains, domain)
		}
	}
	domains = util.RemoveDuplicateElement(domains)
	return domains, nil
}

func (rm *RoleManager) hasAnyRole(name string, domain string) bool {
	patternDomain := rm.getPatternDomain(domain)
	for _, domain := range patternDomain {
		if rm.roles.hasRole(domain, name, rm.matchingFunc) {
			return true
		}
	}
	return false
}

// Roles represents all roles in a domain
type Roles struct {
	sync.Map
}

func (roles *Roles) hasRole(domain, name string, matchingFunc MatchingFunc) bool {
	var ok bool
	if matchingFunc != nil {
		roles.Range(func(key, value interface{}) bool {
			domainPattern, namePattern := getNameAndDomain(key.(string))
			if domainPattern == domain && matchingFunc(name, namePattern) {
				ok = true
			}
			return true
		})
	} else {
		_, ok = roles.Load(getNameWithDomain(domain, name))
	}

	return ok
}

func (roles *Roles) createRole(name string) *Role {
	role, _ := roles.LoadOrStore(name, newRole(name))
	return role.(*Role)
}

// Role represents the data structure for a role in RBAC.
type Role struct {
	nameWithDomain string
	roles          []*Role
}

func newRole(name string) *Role {
	r := Role{}
	r.nameWithDomain = name
	return &r
}

func (r *Role) addRole(role *Role) {
	// determine whether this role has been added
	for _, rr := range r.roles {
		if rr.nameWithDomain == role.nameWithDomain {
			return
		}
	}

	r.roles = append(r.roles, role)
}

func (r *Role) deleteRole(role *Role) {
	for i, rr := range r.roles {
		if rr.nameWithDomain == role.nameWithDomain {
			r.roles = append(r.roles[:i], r.roles[i+1:]...)
			return
		}
	}
}

func (r *Role) hasRole(nameWithDomain string, hierarchyLevel int) bool {
	if r.hasDirectRole(nameWithDomain) {
		return true
	}

	if hierarchyLevel <= 0 {
		return false
	}

	for _, role := range r.roles {
		if role.hasRole(nameWithDomain, hierarchyLevel-1) {
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

func (r *Role) hasDirectRole(nameWithDomain string) bool {
	for _, role := range r.roles {
		if role.nameWithDomain == nameWithDomain {
			return true
		}
	}

	return false
}

func (r *Role) hasDirectRoleWithMatchingFunc(name string, matchingFunc MatchingFunc) bool {
	for _, role := range r.roles {
		_, roleName := getNameAndDomain(role.nameWithDomain)
		if roleName == name || (matchingFunc(name, roleName) && name != roleName) {
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
	sb.WriteString(r.nameWithDomain)
	sb.WriteString(" < ")
	if len(r.roles) != 1 {
		sb.WriteString("(")
	}

	for i, role := range r.roles {
		if i == 0 {
			sb.WriteString(role.nameWithDomain)
		} else {
			sb.WriteString(", ")
			sb.WriteString(role.nameWithDomain)
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
		_, roleName := getNameAndDomain(role.nameWithDomain)
		names = append(names, roleName)
	}
	return names
}

func getNameWithDomain(domain, name string) string {
	if domain == "" {
		return name
	}
	return fmt.Sprintf("%s%s%s", domain, defaultSeparator, name)
}

func getNameAndDomain(domainAndName string) (string, string) {
	t := strings.Split(domainAndName, defaultSeparator)
	if len(t) == 1 {
		return defaultDomain, t[0]
	}
	return t[0], t[1]
}
