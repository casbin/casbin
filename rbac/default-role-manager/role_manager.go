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

type MatchingFunc func(arg1 string, arg2 string) bool

type IPatternMatcher interface {
	IsPattern(str string) bool
	Match(str, pattern string) bool
}

type PatternMatcher struct {
	patternFn func(str string) bool
	matchFn   MatchingFunc
}

func NewPatternMatcher(isPatternFn func(str string) bool, matchFn MatchingFunc) *PatternMatcher {
	return &PatternMatcher{isPatternFn, matchFn}
}

func (m *PatternMatcher) IsPattern(str string) bool {
	return m.patternFn(str)
}
func (m *PatternMatcher) Match(str, pattern string) bool {
	return m.matchFn(str, pattern)
}

type PrefixMatcher struct {
	prefix  string
	matchFn MatchingFunc
}

func NewPrefixMatcher(prefix string, matchFn MatchingFunc) *PrefixMatcher {
	return &PrefixMatcher{prefix, matchFn}
}

func (m *PrefixMatcher) IsPattern(str string) bool {
	return strings.HasPrefix(str, m.prefix)
}
func (m *PrefixMatcher) Match(str, pattern string) bool {
	if m.IsPattern(pattern) {
		return m.matchFn(str, pattern[len(m.prefix):])
	}
	return false
}

// Role represents the data structure for a role in RBAC.
type Role struct {
	name      string
	roles     *sync.Map
	users     *sync.Map
	matched   *sync.Map
	matchedBy *sync.Map
}

func newRole(name string) *Role {
	r := Role{}
	r.name = name
	r.roles = &sync.Map{}
	r.users = &sync.Map{}
	r.matched = &sync.Map{}
	r.matchedBy = &sync.Map{}
	return &r
}

func (r *Role) addRole(role *Role) {
	r.roles.Store(role.name, role)
	role.addUser(r)
}

func (r *Role) removeRole(role *Role) {
	r.roles.Delete(role.name)
	role.removeUser(r)
}

//should only be called inside addRole
func (r *Role) addUser(user *Role) {
	r.users.Store(user.name, user)
}

//should only be called inside removeRole
func (r *Role) removeUser(user *Role) {
	r.users.Delete(user.name)
}

func (r *Role) addMatch(role *Role) {
	r.matched.Store(role.name, role)
	role.matchedBy.Store(r.name, r)
}

func (r *Role) removeMatch(role *Role) {
	r.matched.Delete(role.name)
	role.matchedBy.Delete(r.name)
}

func (r *Role) removeMatches() {
	r.matched.Range(func(key, value interface{}) bool {
		r.removeMatch(value.(*Role))
		return true
	})
	r.matchedBy.Range(func(key, value interface{}) bool {
		value.(*Role).removeMatch(r)
		return true
	})
}

func (r *Role) rangeRoles(fn func(key, value interface{}) bool) {
	r.roles.Range(fn)
	r.roles.Range(func(key, value interface{}) bool {
		role := value.(*Role)
		role.matched.Range(fn)
		return true
	})
	r.matchedBy.Range(func(key, value interface{}) bool {
		role := value.(*Role)
		role.roles.Range(fn)
		return true
	})
}

func (r *Role) rangeUsers(fn func(key, value interface{}) bool) {
	r.users.Range(fn)
	r.users.Range(func(key, value interface{}) bool {
		role := value.(*Role)
		role.matched.Range(fn)
		return true
	})
	r.matchedBy.Range(func(key, value interface{}) bool {
		role := value.(*Role)
		role.users.Range(fn)
		return true
	})
}

func (r *Role) toString() string {
	roles := r.getRoles()

	if len(roles) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(r.name)
	sb.WriteString(" < ")
	if len(roles) != 1 {
		sb.WriteString("(")
	}

	for i, role := range roles {
		if i == 0 {
			sb.WriteString(role)
		} else {
			sb.WriteString(", ")
			sb.WriteString(role)
		}
	}

	if len(roles) != 1 {
		sb.WriteString(")")
	}

	return sb.String()
}

func (r *Role) getRoles() []string {
	var names []string
	r.rangeRoles(func(key, value interface{}) bool {
		names = append(names, key.(string))
		return true
	})
	return util.RemoveDuplicateElement(names)
}

func (r *Role) getUsers() []string {
	var names []string
	r.rangeUsers(func(key, value interface{}) bool {
		names = append(names, key.(string))
		return true
	})
	return names
}

// RoleManagerImpl provides a default implementation for the RoleManager interface
type RoleManagerImpl struct {
	allRoles          *sync.Map
	patternRoles      *sync.Map
	maxHierarchyLevel int
	roleMatcher       IPatternMatcher
	logger            log.Logger
	matchingFuncCache *util.SyncLRUCache
}

// NewRoleManagerImpl is the constructor for creating an instance of the
// default RoleManager implementation.
func NewRoleManagerImpl(maxHierarchyLevel int) *RoleManagerImpl {
	rm := RoleManagerImpl{}
	_ = rm.Clear() //init allRoles and matchingFuncCache
	rm.maxHierarchyLevel = maxHierarchyLevel
	rm.SetLogger(&log.DefaultLogger{})
	return &rm
}

// use this constructor to avoid rebuild of AddMatchingFunc
func newRoleManagerWithMatcher(maxHierarchyLevel int, matcher IPatternMatcher) *RoleManagerImpl {
	rm := NewRoleManagerImpl(maxHierarchyLevel)
	rm.roleMatcher = matcher
	return rm
}

// rebuilds role cache
func (rm *RoleManagerImpl) rebuild() {
	roles := rm.allRoles
	_ = rm.Clear()
	rangeLinks(roles, func(name1, name2 string, domain ...string) bool {
		_ = rm.AddLink(name1, name2, domain...)
		return true
	})
}

func (rm *RoleManagerImpl) Match(str string, pattern string) bool {
	cacheKey := strings.Join([]string{str, pattern}, "$$")
	if v, has := rm.matchingFuncCache.Get(cacheKey); has {
		return v.(bool)
	} else {
		var matched bool
		if rm.roleMatcher != nil {
			matched = rm.roleMatcher.Match(str, pattern)
		} else {
			matched = str == pattern
		}
		rm.matchingFuncCache.Put(cacheKey, matched)
		return matched
	}
}

func (rm *RoleManagerImpl) rangeMatchingRoles(pattern string, fn func(role *Role)) {
	rm.allRoles.Range(func(key, value interface{}) bool {
		name := key.(string)
		if name != pattern && rm.Match(name, pattern) {
			fn(value.(*Role))
		}
		return true
	})
}

func (rm *RoleManagerImpl) rangeMatchingPatterns(name string, fn func(role *Role)) {
	rm.patternRoles.Range(func(key, _ interface{}) bool {
		pattern := key.(string)
		if name != pattern && rm.Match(name, pattern) {
			value, _ := rm.allRoles.Load(pattern)
			fn(value.(*Role))
		}
		return true
	})
}

func (rm *RoleManagerImpl) load(name interface{}) (value *Role, ok bool) {
	if r, ok := rm.allRoles.Load(name); ok {
		return r.(*Role), true
	}
	return nil, false
}

// loads or creates a role
func (rm *RoleManagerImpl) getRole(name string) (r *Role, created bool) {
	var role *Role
	var ok bool

	if role, ok = rm.load(name); !ok {
		role = newRole(name)
		rm.allRoles.Store(name, role)

		if rm.roleMatcher != nil {
			if rm.roleMatcher.IsPattern(name) {
				rm.patternRoles.Store(name, nil)
				rm.rangeMatchingRoles(name, func(r *Role) {
					role.addMatch(r)
				})
			}
			rm.rangeMatchingPatterns(name, func(r *Role) {
				r.addMatch(role)
			})
		}
	}

	return role, !ok
}

func loadAndDelete(m *sync.Map, name string) (value interface{}, loaded bool) {
	value, loaded = m.Load(name)
	if loaded {
		m.Delete(name)
	}
	return value, loaded
}

func (rm *RoleManagerImpl) removeRole(name string) {
	if role, ok := loadAndDelete(rm.allRoles, name); ok {
		role.(*Role).removeMatches()
	}
	rm.patternRoles.Delete(name)
}

// AddMatchingFunc support use pattern in g
func (rm *RoleManagerImpl) AddMatchingFunc(name string, fn MatchingFunc) {
	matcher := NewPatternMatcher(func(str string) bool {
		return true
	}, fn)
	rm.SetRoleMatcher(matcher)
}

func (rm *RoleManagerImpl) SetRoleMatcher(matcher IPatternMatcher) {
	rm.roleMatcher = matcher
	rm.rebuild()
}

// AddDomainMatchingFunc support use domain pattern in g
func (rm *RoleManagerImpl) AddDomainMatchingFunc(name string, fn MatchingFunc) {}

// SetLogger sets role manager's logger.
func (rm *RoleManagerImpl) SetLogger(logger log.Logger) {
	rm.logger = logger
}

// Clear clears all stored data and resets the role manager to the initial state.
func (rm *RoleManagerImpl) Clear() error {
	rm.matchingFuncCache = util.NewSyncLRUCache(100)
	rm.allRoles = &sync.Map{}
	rm.patternRoles = &sync.Map{}
	return nil
}

// AddLink adds the inheritance link between role: name1 and role: name2.
// aka role: name1 inherits role: name2.
func (rm *RoleManagerImpl) AddLink(name1 string, name2 string, domains ...string) error {
	user, _ := rm.getRole(name1)
	role, _ := rm.getRole(name2)
	user.addRole(role)
	return nil
}

// DeleteLink deletes the inheritance link between role: name1 and role: name2.
// aka role: name1 does not inherit role: name2 any more.
func (rm *RoleManagerImpl) DeleteLink(name1 string, name2 string, domains ...string) error {
	user, _ := rm.getRole(name1)
	role, _ := rm.getRole(name2)
	user.removeRole(role)
	return nil
}

// HasLink determines whether role: name1 inherits role: name2.
func (rm *RoleManagerImpl) HasLink(name1 string, name2 string, domains ...string) (bool, error) {
	if name1 == name2 || (rm.roleMatcher != nil && rm.Match(name1, name2)) {
		return true, nil
	}

	user, userCreated := rm.getRole(name1)
	role, roleCreated := rm.getRole(name2)

	if userCreated {
		defer rm.removeRole(user.name)
	}
	if roleCreated {
		defer rm.removeRole(role.name)
	}

	return rm.hasLinkHelper(role.name, map[string]*Role{user.name: user}, rm.maxHierarchyLevel), nil
}

func (rm *RoleManagerImpl) hasLinkHelper(targetName string, roles map[string]*Role, level int) bool {
	if level < 0 || len(roles) == 0 {
		return false
	}

	nextRoles := map[string]*Role{}
	for _, role := range roles {
		if targetName == role.name || (rm.roleMatcher != nil && rm.Match(role.name, targetName)) {
			return true
		}
		role.rangeRoles(func(key, value interface{}) bool {
			nextRoles[key.(string)] = value.(*Role)
			return true
		})
	}

	return rm.hasLinkHelper(targetName, nextRoles, level-1)
}

// GetRoles gets the roles that a user inherits.
func (rm *RoleManagerImpl) GetRoles(name string, domains ...string) ([]string, error) {
	user, created := rm.getRole(name)
	if created {
		defer rm.removeRole(user.name)
	}
	return user.getRoles(), nil
}

// GetUsers gets the users of a role.
// domain is an unreferenced parameter here, may be used in other implementations.
func (rm *RoleManagerImpl) GetUsers(name string, domain ...string) ([]string, error) {
	role, created := rm.getRole(name)
	if created {
		defer rm.removeRole(role.name)
	}
	return role.getUsers(), nil
}

func (rm *RoleManagerImpl) toString() []string {
	var roles []string

	rm.allRoles.Range(func(key, value interface{}) bool {
		role := value.(*Role)
		if text := role.toString(); text != "" {
			roles = append(roles, text)
		}
		return true
	})

	return roles
}

// PrintRoles prints all the roles to log.
func (rm *RoleManagerImpl) PrintRoles() error {
	if !(rm.logger).IsEnabled() {
		return nil
	}
	roles := rm.toString()
	rm.logger.LogRole(roles)
	return nil
}

// GetDomains gets domains that a user has
func (rm *RoleManagerImpl) GetDomains(name string) ([]string, error) {
	domains := []string{defaultDomain}
	return domains, nil
}

// GetAllDomains gets all domains
func (rm *RoleManagerImpl) GetAllDomains() ([]string, error) {
	domains := []string{defaultDomain}
	return domains, nil
}

func rangeLinks(users *sync.Map, fn func(name1, name2 string, domain ...string) bool) {
	users.Range(func(_, value interface{}) bool {
		user := value.(*Role)
		user.roles.Range(func(key, _ interface{}) bool {
			roleName := key.(string)
			return fn(user.name, roleName, defaultDomain)
		})
		return true
	})
}

func (rm *RoleManagerImpl) Range(fn func(name1, name2 string, domain ...string) bool) {
	rangeLinks(rm.allRoles, fn)
}

// Deprecated: BuildRelationship is no longer required
func (rm *RoleManagerImpl) BuildRelationship(name1 string, name2 string, domain ...string) error {
	return nil
}

type DomainManager struct {
	rmMap             *sync.Map
	patternMap        *sync.Map
	maxHierarchyLevel int
	roleMatcher       IPatternMatcher
	domainMatcher     IPatternMatcher
	logger            log.Logger
	matchingFuncCache *util.SyncLRUCache
}

// NewDomainManager is the constructor for creating an instance of the
// default DomainManager implementation.
func NewDomainManager(maxHierarchyLevel int) *DomainManager {
	dm := &DomainManager{}
	_ = dm.Clear() // init rmMap and rmCache
	dm.maxHierarchyLevel = maxHierarchyLevel
	return dm
}

// SetLogger sets role manager's logger.
func (dm *DomainManager) SetLogger(logger log.Logger) {
	dm.logger = logger
}

// AddMatchingFunc support use pattern in g
func (dm *DomainManager) AddMatchingFunc(name string, fn MatchingFunc) {
	matcher := NewPatternMatcher(func(str string) bool {
		return true
	}, fn)
	dm.SetRoleMatcher(matcher)
}

func (dm *DomainManager) SetRoleMatcher(matcher IPatternMatcher) {
	dm.roleMatcher = matcher
	dm.rmMap.Range(func(key, value interface{}) bool {
		value.(*RoleManagerImpl).SetRoleMatcher(matcher)
		return true
	})
}

// AddDomainMatchingFunc support use domain pattern in g
func (dm *DomainManager) AddDomainMatchingFunc(name string, fn MatchingFunc) {
	matcher := NewPatternMatcher(func(str string) bool {
		return true
	}, fn)
	dm.SetDomainMatcher(matcher)
}

func (dm *DomainManager) SetDomainMatcher(matcher IPatternMatcher) {
	dm.domainMatcher = matcher
	dm.rebuild()
}

// clears the map of RoleManagers
func (dm *DomainManager) rebuild() {
	rmMap := dm.rmMap
	_ = dm.Clear()
	rmMap.Range(func(key, value interface{}) bool {
		domain := key.(string)
		rm := value.(*RoleManagerImpl)

		rm.Range(func(name1, name2 string, _ ...string) bool {
			_ = dm.AddLink(name1, name2, domain)
			return true
		})
		return true
	})
}

//Clear clears all stored data and resets the role manager to the initial state.
func (dm *DomainManager) Clear() error {
	dm.rmMap = &sync.Map{}
	dm.patternMap = &sync.Map{}
	dm.matchingFuncCache = util.NewSyncLRUCache(100)
	return nil
}

func (dm *DomainManager) getDomain(domains ...string) (domain string, err error) {
	switch len(domains) {
	case 0:
		return defaultDomain, nil
	case 1:
		return domains[0], nil
	default:
		return "", errors.ERR_DOMAIN_PARAMETER
	}
}

func (dm *DomainManager) Match(str string, pattern string) bool {
	cacheKey := strings.Join([]string{str, pattern}, "$$")
	if v, has := dm.matchingFuncCache.Get(cacheKey); has {
		return v.(bool)
	} else {
		var matched bool
		if dm.domainMatcher != nil {
			matched = dm.domainMatcher.Match(str, pattern)
		} else {
			matched = str == pattern
		}
		dm.matchingFuncCache.Put(cacheKey, matched)
		return matched
	}
}

func (dm *DomainManager) rangeAffectedRoleManagers(pattern string, fn func(rm *RoleManagerImpl)) {
	dm.rmMap.Range(func(key, value interface{}) bool {
		domain := key.(string)
		if domain != pattern && dm.Match(domain, pattern) {
			fn(value.(*RoleManagerImpl))
		}
		return true
	})
}

func (dm *DomainManager) rangeMatchingPatterns(domain string, fn func(rm *RoleManagerImpl)) {
	dm.patternMap.Range(func(key, _ interface{}) bool {
		pattern := key.(string)
		if domain != pattern && dm.Match(domain, pattern) {
			value, _ := dm.load(pattern)
			fn(value)
		}
		return true
	})
}

func (dm *DomainManager) load(name interface{}) (value *RoleManagerImpl, ok bool) {
	if r, ok := dm.rmMap.Load(name); ok {
		return r.(*RoleManagerImpl), true
	}
	return nil, false
}

// load or create a RoleManager instance of domain
func (dm *DomainManager) getRoleManager(domain string, store bool) *RoleManagerImpl {
	var rm *RoleManagerImpl
	var ok bool

	if rm, ok = dm.load(domain); !ok {
		rm = newRoleManagerWithMatcher(dm.maxHierarchyLevel, dm.roleMatcher)
		if store {
			dm.rmMap.Store(domain, rm)
		}
		if dm.domainMatcher != nil {
			if dm.domainMatcher.IsPattern(domain) {
				dm.patternMap.Store(domain, nil)
			}
			dm.rangeMatchingPatterns(domain, func(rm2 *RoleManagerImpl) {
				rm2.Range(func(name1, name2 string, domain ...string) bool {
					_ = rm.AddLink(name1, name2, domain...)
					return true
				})
			})
		}
	}
	return rm
}

// AddLink adds the inheritance link between role: name1 and role: name2.
// aka role: name1 inherits role: name2.
func (dm *DomainManager) AddLink(name1 string, name2 string, domains ...string) error {
	domain, err := dm.getDomain(domains...)
	if err != nil {
		return err
	}
	roleManager := dm.getRoleManager(domain, true) //create role manager if it does not exist
	_ = roleManager.AddLink(name1, name2, domains...)

	if dm.domainMatcher != nil && dm.domainMatcher.IsPattern(domain) {
		dm.rangeAffectedRoleManagers(domain, func(rm *RoleManagerImpl) {
			_ = rm.AddLink(name1, name2, domains...)
		})
	}
	return nil
}

// DeleteLink deletes the inheritance link between role: name1 and role: name2.
// aka role: name1 does not inherit role: name2 any more.
func (dm *DomainManager) DeleteLink(name1 string, name2 string, domains ...string) error {
	domain, err := dm.getDomain(domains...)
	if err != nil {
		return err
	}
	roleManager := dm.getRoleManager(domain, true) //create role manager if it does not exist
	_ = roleManager.DeleteLink(name1, name2, domains...)

	if dm.domainMatcher != nil && dm.domainMatcher.IsPattern(domain) {
		dm.rangeAffectedRoleManagers(domain, func(rm *RoleManagerImpl) {
			_ = rm.DeleteLink(name1, name2, domains...)
		})
	}
	return nil
}

// HasLink determines whether role: name1 inherits role: name2.
func (dm *DomainManager) HasLink(name1 string, name2 string, domains ...string) (bool, error) {
	domain, err := dm.getDomain(domains...)
	if err != nil {
		return false, err
	}
	rm := dm.getRoleManager(domain, false)
	return rm.HasLink(name1, name2, domains...)
}

// GetRoles gets the roles that a subject inherits.
func (dm *DomainManager) GetRoles(name string, domains ...string) ([]string, error) {
	domain, err := dm.getDomain(domains...)
	if err != nil {
		return nil, err
	}
	rm := dm.getRoleManager(domain, false)
	return rm.GetRoles(name, domains...)
}

// GetUsers gets the users of a role.
func (dm *DomainManager) GetUsers(name string, domains ...string) ([]string, error) {
	domain, err := dm.getDomain(domains...)
	if err != nil {
		return nil, err
	}
	rm := dm.getRoleManager(domain, false)
	return rm.GetUsers(name, domains...)
}

func (dm *DomainManager) toString() []string {
	var roles []string

	dm.rmMap.Range(func(key, value interface{}) bool {
		domain := key.(string)
		rm := value.(*RoleManagerImpl)
		domainRoles := rm.toString()
		roles = append(roles, fmt.Sprintf("%s: %s", domain, strings.Join(domainRoles, ", ")))
		return true
	})

	return roles
}

// PrintRoles prints all the roles to log.
func (dm *DomainManager) PrintRoles() error {
	if !(dm.logger).IsEnabled() {
		return nil
	}

	roles := dm.toString()
	dm.logger.LogRole(roles)
	return nil
}

// GetDomains gets domains that a user has
func (dm *DomainManager) GetDomains(name string) ([]string, error) {
	var domains []string
	dm.rmMap.Range(func(key, value interface{}) bool {
		domain := key.(string)
		rm := value.(*RoleManagerImpl)
		role, created := rm.getRole(name)
		if created {
			defer rm.removeRole(role.name)
		}
		if len(role.getUsers()) > 0 || len(role.getRoles()) > 0 {
			domains = append(domains, domain)
		}
		return true
	})
	return domains, nil
}

// GetAllDomains gets all domains
func (rm *DomainManager) GetAllDomains() ([]string, error) {
	var domains []string
	rm.rmMap.Range(func(key, value interface{}) bool {
		domains = append(domains, key.(string))
		return true
	})
	return domains, nil
}

// Deprecated: BuildRelationship is no longer required
func (rm *DomainManager) BuildRelationship(name1 string, name2 string, domain ...string) error {
	return nil
}

type RoleManager struct {
	*DomainManager
}

func NewRoleManager(maxHierarchyLevel int) *RoleManager {
	rm := &RoleManager{}
	rm.DomainManager = NewDomainManager(maxHierarchyLevel)
	return rm
}
