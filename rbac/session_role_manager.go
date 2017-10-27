// Copyright 2017 EDOMO Systems GmbH. All Rights Reserved.
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

package rbac

import (
	"sort"
	
	"github.com/casbin/casbin/util"
)

type sessionRoleManager struct {
	allRoles          map[string]*SessionRole
	maxHierarchyLevel int
}

// SessionRoleManager provides an implementation for the RoleManagerConstructor that
// supports RBAC sessions with a start time and an end time.
func SessionRoleManager() RoleManagerConstructor {
	return func() RoleManager {
		return NewSessionRoleManager(10)
	}
}

// NewSessionRoleManager is the constructor for creating an instance of the
// SessionRoleManager implementation.
func NewSessionRoleManager(maxHierarchyLevel int) RoleManager {
	rm := sessionRoleManager{}
	rm.allRoles = make(map[string]*SessionRole)
	rm.maxHierarchyLevel = maxHierarchyLevel
	return &rm
}

func (rm *sessionRoleManager) hasRole(name string) bool {
	_, ok := rm.allRoles[name]
	return ok
}

func (rm *sessionRoleManager) createRole(name string) *SessionRole {
	if !rm.hasRole(name) {
		rm.allRoles[name] = newSessionRole(name)
	}
	return rm.allRoles[name]
}

func (rm *sessionRoleManager) AddLink(name1 string, name2 string, timeRange ...string) {
	if len(timeRange) != 2 {
		return
	}
	startTime := timeRange[0]
	endTime := timeRange[1]

	role1 := rm.createRole(name1)
	role2 := rm.createRole(name2)

	session := Session{role2, startTime, endTime}
	role1.addSession(session)
}

func (rm *sessionRoleManager) DeleteLink(name1 string, name2 string, unused ...string) {
	if !rm.hasRole(name1) || !rm.hasRole(name2) {
		return
	}

	role1 := rm.createRole(name1)
	role2 := rm.createRole(name2)

	role1.deleteSessions(role2.name)
}

func (rm *sessionRoleManager) HasLink(name1 string, name2 string, requestTime ...string) bool {
	if len(requestTime) != 1 {
		return false
	}

	if name1 == name2 {
		return true
	}

	if !rm.hasRole(name1) || !rm.hasRole(name2) {
		return false
	}

	role1 := rm.createRole(name1)
	return role1.hasValidSession(name2, rm.maxHierarchyLevel, requestTime[0])
}

func (rm *sessionRoleManager) GetRoles(name string, currentTime ...string) []string {
	if len(currentTime) != 1 {
		return nil
	}
	requestTime := currentTime[0]

	if !rm.hasRole(name) {
		return nil
	}

	sessionRoles := rm.createRole(name).getSessionRoles(requestTime)
	return sessionRoles
}

func (rm *sessionRoleManager) GetUsers(name string, currentTime ...string) []string {
	if len(currentTime) != 1 {
		return nil
	}
	requestTime := currentTime[0]

	users := []string{}
	for _, role := range rm.allRoles {
		if role.hasDirectRole(name, requestTime) {
			users = append(users, role.name)
		}
	}
	sort.Strings(users)
	return users
}

func (rm *sessionRoleManager) PrintRoles() {
	for _, role := range rm.allRoles {
		util.LogPrint(role.toString())
	}
}

// SessionRole is a modified version of the default role.
// A SessionRole not only has a name, but also a list of sessions.
type SessionRole struct {
	name     string
	sessions []Session
}

func newSessionRole(name string) *SessionRole {
	sr := SessionRole{name: name}
	return &sr
}

func (sr *SessionRole) addSession(s Session) {
	sr.sessions = append(sr.sessions, s)
}

func (sr *SessionRole) deleteSessions(sessionName string) {
	// Delete sessions from an array while iterating it
	index := 0
	for _, srs := range sr.sessions {
		if srs.role.name != sessionName {
			sr.sessions[index] = srs
			index++
		}
	}
	sr.sessions = sr.sessions[:index]
}

//
//func (sr *SessionRole) getSessions() []Session {
//	return sr.sessions
//}

func (sr *SessionRole) getSessionRoles(requestTime string) []string {
	names := []string{}
	for _, session := range sr.sessions {
		if session.startTime <= requestTime && requestTime <= session.endTime {
			if !contains(names, session.role.name) {
				names = append(names, session.role.name)
			}
		}
	}
	return names
}

func (sr *SessionRole) hasValidSession(name string, hierarchyLevel int, requestTime string) bool {
	if hierarchyLevel == 1 {
		return sr.name == name
	}

	for _, s := range sr.sessions {
		if s.startTime <= requestTime && requestTime <= s.endTime {
			if s.role.name == name {
				return true
			}
			if s.role.hasValidSession(name, hierarchyLevel-1, requestTime) {
				return true
			}
		}
	}
	return false
}

func (sr *SessionRole) hasDirectRole(name string, requestTime string) bool {
	for _, session := range sr.sessions {
		if session.role.name == name {
			if session.startTime <= requestTime && requestTime <= session.endTime {
				return true
			}
		}
	}
	return false
}

func (sr *SessionRole) toString() string {
	sessions := ""
	for i, session := range sr.sessions {
		if i == 0 {
			sessions += session.role.name
		} else {
			sessions += ", " + session.role.name
		}
		sessions += " (until: " + session.endTime + ")"
	}
	return sr.name + " < " + sessions
}

// Session represents the activation of a role inheritance for a
// specified time. A role inheritance is always bound to its temporal validity.
// As soon as a session loses its validity, the corresponding role inheritance
// becomes invalid too.
type Session struct {
	role      *SessionRole
	startTime string
	endTime   string
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
