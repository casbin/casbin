package rbac

import (
	"log"
)

// RoleManager is the interface to manage the roles in RBAC.
type RoleManager struct {
	allRoles map[string]*Role
	level    int
}

// The constructor for RoleManager.
func NewRoleManager(level int) *RoleManager {
	rm := RoleManager{}
	rm.allRoles = make(map[string]*Role)
	rm.level = level
	return &rm
}

func (rm *RoleManager) hasRole(name string) bool {
	_, ok := rm.allRoles[name]
	return ok
}

func (rm *RoleManager) createRole(name string) *Role {
	if !rm.hasRole(name) {
		rm.allRoles[name] = newRole(name)
	}

	return rm.allRoles[name]
}

// Add the link between role: name1 and role: name2.
// aka name1 inherits role: name2.
func (rm *RoleManager) AddLink(name1 string, name2 string) {
	role1 := rm.createRole(name1)
	role2 := rm.createRole(name2)
	role1.addRole(role2)
}

// Whether role: name1 inherits role: name2.
func (rm *RoleManager) HasLink(name1 string, name2 string) bool {
	if name1 == name2 {
		return true
	}

	if !rm.hasRole(name1) || !rm.hasRole(name2) {
		return false
	}

	role1 := rm.createRole(name1)
	return role1.hasRole(name2, rm.level)
}

// Get the roles that a subject inherits.
func (rm *RoleManager) GetRoles(name string) []string {
	if rm.hasRole(name) {
		return rm.createRole(name).getRoles()
	} else {
		return nil
	}
}

// Print all the roles.
func (rm *RoleManager) PrintRoles() {
	for _, role := range rm.allRoles {
		log.Print(role.toString())
	}
}

// Role is the data structure for a role in RBAC.
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

func (r *Role) hasRole(name string, level int) bool {
	if r.name == name {
		return true
	}

	if level <= 0 {
		return false
	}

	for _, role := range r.roles {
		if role.hasRole(name, level-1) {
			return true
		}
	}
	return false
}

func (r *Role) toString() string {
	names := ""
	for i, role := range r.roles {
		if i == 0 {
			names += role.name
		} else {
			names += ", " + role.name
		}
	}
	return r.name + " < " + names
}

func (r *Role) getRoles() []string {
	names := []string{}
	for _, role := range r.roles {
		names = append(names, role.name)
	}
	return names
}
