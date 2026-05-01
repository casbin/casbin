package defaultrolemanager

import (
	"sync"

	"github.com/casbin/casbin/v3/rbac"
	"github.com/casbin/casbin/v3/util"
)



type RoleManagerImpl struct {
	allRoles          *sync.Map
	maxHierarchyLevel int
	matchingFunc      rbac.MatchingFunc
	matchingFuncCache *util.SyncLRUCache
}

func NewRoleManager(maxHierarchyLevel int) rbac.RoleManager {
	return NewRoleManagerImpl(maxHierarchyLevel)
}

func NewRoleManagerImpl(maxHierarchyLevel int) *RoleManagerImpl {
	rm := &RoleManagerImpl{}
	_ = rm.Clear()
	rm.maxHierarchyLevel = maxHierarchyLevel
	return rm
}

func NewConditionalRoleManager(maxHierarchyLevel int) rbac.ConditionalRoleManager {
	return NewRoleManagerImpl(maxHierarchyLevel)
}


func (rm *RoleManagerImpl) Clear() error                                          { rm.allRoles = &sync.Map{}; rm.matchingFuncCache = util.NewSyncLRUCache(100); return nil }
func (rm *RoleManagerImpl) BuildRelationship(n1, n2 string, d ...string) error   { return nil }
func (rm *RoleManagerImpl) AddLink(n1, n2 string, d ...string) error             { return nil }
func (rm *RoleManagerImpl) DeleteLink(n1, n2 string, d ...string) error          { return nil }
func (rm *RoleManagerImpl) HasLink(n1, n2 string, d ...string) (bool, error)     { return false, nil }
func (rm *RoleManagerImpl) GetRoles(n string, d ...string) ([]string, error)     { return nil, nil }
func (rm *RoleManagerImpl) GetUsers(n string, d ...string) ([]string, error)     { return nil, nil }
func (rm *RoleManagerImpl) GetImplicitRoles(n string, d ...string) ([]string, error) { return nil, nil }
func (rm *RoleManagerImpl) GetImplicitUsers(n string, d ...string) ([]string, error) { return nil, nil }
func (rm *RoleManagerImpl) GetDomains(name string) ([]string, error)             { return nil, nil }
func (rm *RoleManagerImpl) GetAllDomains() ([]string, error)                     { return nil, nil }
func (rm *RoleManagerImpl) PrintRoles() error                                    { return nil }
func (rm *RoleManagerImpl) Match(n1, n2 string) bool                             { return false }
func (rm *RoleManagerImpl) AddMatchingFunc(n string, fn rbac.MatchingFunc)       { rm.matchingFunc = fn }
func (rm *RoleManagerImpl) AddDomainMatchingFunc(n string, fn rbac.MatchingFunc) {}
func (rm *RoleManagerImpl) DeleteDomain(domain string) error                     { return nil }


func (rm *RoleManagerImpl) AddLinkConditionFunc(user, role string, fn rbac.LinkConditionFunc)                    {}
func (rm *RoleManagerImpl) SetLinkConditionFuncParams(user, role string, params ...string)                       {}
func (rm *RoleManagerImpl) AddDomainLinkConditionFunc(user, role, domain string, fn rbac.LinkConditionFunc)      {}
func (rm *RoleManagerImpl) SetDomainLinkConditionFuncParams(user, role, domain string, params ...string)         {}

func (rm *RoleManagerImpl) Range(fn func(name1, name2 string, domain ...string) bool) {
	rm.allRoles.Range(func(key, value interface{}) bool {
		return true
	})
}



type DomainManager struct {
	rmMap              *sync.Map
	maxHierarchyLevel  int
	matchingFunc       rbac.MatchingFunc
	domainMatchingFunc rbac.MatchingFunc
	matchingFuncCache  *util.SyncLRUCache
}

func NewDomainManager(maxHierarchyLevel int) rbac.RoleManager {
	dm := &DomainManager{}
	_ = dm.Clear()
	dm.maxHierarchyLevel = maxHierarchyLevel
	return dm
}

func NewConditionalDomainManager(maxHierarchyLevel int) rbac.ConditionalRoleManager {
	dm := &DomainManager{}
	_ = dm.Clear()
	dm.maxHierarchyLevel = maxHierarchyLevel
	return dm
}


func (dm *DomainManager) Clear() error                                          { dm.rmMap = &sync.Map{}; dm.matchingFuncCache = util.NewSyncLRUCache(100); return nil }
func (dm *DomainManager) BuildRelationship(n1, n2 string, d ...string) error   { return nil }
func (dm *DomainManager) AddLink(n1, n2 string, d ...string) error             { return nil }
func (dm *DomainManager) DeleteLink(n1, n2 string, d ...string) error          { return nil }
func (dm *DomainManager) HasLink(n1, n2 string, d ...string) (bool, error)     { return false, nil }
func (dm *DomainManager) GetRoles(n string, d ...string) ([]string, error)     { return nil, nil }
func (dm *DomainManager) GetUsers(n string, d ...string) ([]string, error)     { return nil, nil }
func (dm *DomainManager) GetImplicitRoles(n string, d ...string) ([]string, error) { return nil, nil }
func (dm *DomainManager) GetImplicitUsers(n string, d ...string) ([]string, error) { return nil, nil }
func (dm *DomainManager) GetDomains(name string) ([]string, error)             { return nil, nil }
func (dm *DomainManager) GetAllDomains() ([]string, error)                     { return nil, nil }
func (dm *DomainManager) PrintRoles() error                                    { return nil }
func (dm *DomainManager) Match(n1, n2 string) bool                             { return false }
func (dm *DomainManager) AddMatchingFunc(n string, fn rbac.MatchingFunc)       { dm.matchingFunc = fn }
func (dm *DomainManager) AddDomainMatchingFunc(n string, fn rbac.MatchingFunc) { dm.domainMatchingFunc = fn }
func (dm *DomainManager) DeleteDomain(domain string) error                     { return nil }


func (dm *DomainManager) AddLinkConditionFunc(user, role string, fn rbac.LinkConditionFunc)                    {}
func (dm *DomainManager) SetLinkConditionFuncParams(user, role string, params ...string)                       {}
func (dm *DomainManager) AddDomainLinkConditionFunc(user, role, domain string, fn rbac.LinkConditionFunc)      {}
func (dm *DomainManager) SetDomainLinkConditionFuncParams(user, role, domain string, params ...string)         {}
