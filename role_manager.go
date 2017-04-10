package main

type RoleManager struct {
	allRoles map[string]*Role
	level int
}

func newRoleManager(level int) (*RoleManager) {
	rm := RoleManager{}
	rm.allRoles = make(map[string]*Role)
	rm.level = level
	return &rm
}

func (rm *RoleManager) hasRole(name string) bool {
	_, ok := rm.allRoles[name]
	return ok
}

func (rm *RoleManager) createRole(name string) (*Role) {
	if !rm.hasRole(name) {
		rm.allRoles[name] = newRole(name)
	}

	return rm.allRoles[name]
}

func (rm *RoleManager) addLink(name1 string, name2 string) {
	role1 := rm.createRole(name1)
	role2 := rm.createRole(name2)
	role1.addRole(role2)
}

func (rm *RoleManager) hasLink(name1 string, name2 string) bool {
	if !rm.hasRole(name1) || !rm.hasRole(name2) {
		return false
	}
	role1 := rm.createRole(name1)
	return role1.hasRole(name2, rm.level)
}

type Role struct {
	name string
	roles []*Role
}

func newRole(name string) (*Role) {
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
		if role.hasRole(name, level - 1) {
			return true
		}
	}
	return false
}
