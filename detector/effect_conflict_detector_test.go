// Copyright 2025 The casbin Authors. All Rights Reserved.
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

package detector

import (
	"strings"
	"testing"

	"github.com/casbin/casbin/v3/model"
	defaultrolemanager "github.com/casbin/casbin/v3/rbac/default-role-manager"
)

func TestEffectConflictDetector_NilModel(t *testing.T) {
	detector := NewEffectConflictDetector()
	rm := defaultrolemanager.NewRoleManagerImpl(10)
	
	err := detector.CheckModel(nil, rm)
	if err == nil {
		t.Error("Expected error for nil model, but got nil")
	} else if !strings.Contains(err.Error(), "model cannot be nil") {
		t.Errorf("Expected error message to contain 'model cannot be nil', got: %s", err.Error())
	}
}

func TestEffectConflictDetector_NilRoleManager(t *testing.T) {
	detector := NewEffectConflictDetector()
	m := model.Model{}
	
	err := detector.CheckModel(m, nil)
	if err == nil {
		t.Error("Expected error for nil role manager, but got nil")
	} else if !strings.Contains(err.Error(), "role manager cannot be nil") {
		t.Errorf("Expected error message to contain 'role manager cannot be nil', got: %s", err.Error())
	}
}

func TestEffectConflictDetector_NoConflict(t *testing.T) {
	detector := NewEffectConflictDetector()
	rm := defaultrolemanager.NewRoleManagerImpl(10)
	
	// Create a simple model
	m := model.Model{}
	m.AddDef("p", "p", "sub, obj, act, eft")
	m.AddDef("g", "g", "_, _")
	
	// Add policies - no conflicts
	_ = m.AddPolicy("p", "p", []string{"alice", "data1", "read", "allow"})
	_ = m.AddPolicy("p", "p", []string{"data2_admin", "data2", "write", "allow"})
	
	// Add role assignment
	_ = m.AddPolicy("g", "g", []string{"alice", "data2_admin"})
	
	err := detector.CheckModel(m, rm)
	if err != nil {
		t.Errorf("Expected no conflict, but got error: %v", err)
	}
}

func TestEffectConflictDetector_UserAllowRoleDeny(t *testing.T) {
	detector := NewEffectConflictDetector()
	rm := defaultrolemanager.NewRoleManagerImpl(10)
	
	// Create a model
	m := model.Model{}
	m.AddDef("p", "p", "sub, obj, act, eft")
	m.AddDef("g", "g", "_, _")
	
	// alice is allowed to write to data2
	_ = m.AddPolicy("p", "p", []string{"alice", "data2", "write", "allow"})
	// data2_admin is denied to write to data2
	_ = m.AddPolicy("p", "p", []string{"data2_admin", "data2", "write", "deny"})
	
	// alice has role data2_admin
	_ = m.AddPolicy("g", "g", []string{"alice", "data2_admin"})
	
	err := detector.CheckModel(m, rm)
	if err == nil {
		t.Error("Expected conflict detection error, but got nil")
	} else {
		errMsg := err.Error()
		if !strings.Contains(errMsg, "effect conflict detected") {
			t.Errorf("Expected error message to contain 'effect conflict detected', got: %s", errMsg)
		}
		if !strings.Contains(errMsg, "alice") {
			t.Errorf("Expected error message to contain 'alice', got: %s", errMsg)
		}
		if !strings.Contains(errMsg, "data2_admin") {
			t.Errorf("Expected error message to contain 'data2_admin', got: %s", errMsg)
		}
	}
}

func TestEffectConflictDetector_UserDenyRoleAllow(t *testing.T) {
	detector := NewEffectConflictDetector()
	rm := defaultrolemanager.NewRoleManagerImpl(10)
	
	// Create a model
	m := model.Model{}
	m.AddDef("p", "p", "sub, obj, act, eft")
	m.AddDef("g", "g", "_, _")
	
	// alice is denied to read data1
	_ = m.AddPolicy("p", "p", []string{"alice", "data1", "read", "deny"})
	// admin is allowed to read data1
	_ = m.AddPolicy("p", "p", []string{"admin", "data1", "read", "allow"})
	
	// alice has role admin
	_ = m.AddPolicy("g", "g", []string{"alice", "admin"})
	
	err := detector.CheckModel(m, rm)
	if err == nil {
		t.Error("Expected conflict detection error, but got nil")
	} else {
		errMsg := err.Error()
		if !strings.Contains(errMsg, "effect conflict detected") {
			t.Errorf("Expected error message to contain 'effect conflict detected', got: %s", errMsg)
		}
		if !strings.Contains(errMsg, "alice") {
			t.Errorf("Expected error message to contain 'alice', got: %s", errMsg)
		}
		if !strings.Contains(errMsg, "admin") {
			t.Errorf("Expected error message to contain 'admin', got: %s", errMsg)
		}
	}
}

func TestEffectConflictDetector_NoRoles(t *testing.T) {
	detector := NewEffectConflictDetector()
	rm := defaultrolemanager.NewRoleManagerImpl(10)
	
	// Create a model
	m := model.Model{}
	m.AddDef("p", "p", "sub, obj, act, eft")
	m.AddDef("g", "g", "_, _")
	
	// Add policies but no role assignments
	_ = m.AddPolicy("p", "p", []string{"alice", "data1", "read", "allow"})
	_ = m.AddPolicy("p", "p", []string{"bob", "data2", "write", "deny"})
	
	err := detector.CheckModel(m, rm)
	if err != nil {
		t.Errorf("Expected no conflict when there are no role assignments, but got error: %v", err)
	}
}

func TestEffectConflictDetector_SameEffect(t *testing.T) {
	detector := NewEffectConflictDetector()
	rm := defaultrolemanager.NewRoleManagerImpl(10)
	
	// Create a model
	m := model.Model{}
	m.AddDef("p", "p", "sub, obj, act, eft")
	m.AddDef("g", "g", "_, _")
	
	// Both alice and admin are allowed to read data1
	_ = m.AddPolicy("p", "p", []string{"alice", "data1", "read", "allow"})
	_ = m.AddPolicy("p", "p", []string{"admin", "data1", "read", "allow"})
	
	// alice has role admin
	_ = m.AddPolicy("g", "g", []string{"alice", "admin"})
	
	err := detector.CheckModel(m, rm)
	if err != nil {
		t.Errorf("Expected no conflict when effects are the same, but got error: %v", err)
	}
}

func TestEffectConflictDetector_MultipleRoles(t *testing.T) {
	detector := NewEffectConflictDetector()
	rm := defaultrolemanager.NewRoleManagerImpl(10)
	
	// Create a model
	m := model.Model{}
	m.AddDef("p", "p", "sub, obj, act, eft")
	m.AddDef("g", "g", "_, _")
	
	// alice is allowed to write to data1
	_ = m.AddPolicy("p", "p", []string{"alice", "data1", "write", "allow"})
	// admin is allowed to write to data1
	_ = m.AddPolicy("p", "p", []string{"admin", "data1", "write", "allow"})
	// moderator is denied to write to data1
	_ = m.AddPolicy("p", "p", []string{"moderator", "data1", "write", "deny"})
	
	// alice has both admin and moderator roles
	_ = m.AddPolicy("g", "g", []string{"alice", "admin"})
	_ = m.AddPolicy("g", "g", []string{"alice", "moderator"})
	
	err := detector.CheckModel(m, rm)
	if err == nil {
		t.Error("Expected conflict detection error with multiple roles, but got nil")
	} else {
		errMsg := err.Error()
		if !strings.Contains(errMsg, "effect conflict detected") {
			t.Errorf("Expected error message to contain 'effect conflict detected', got: %s", errMsg)
		}
	}
}

func TestEffectConflictDetector_DefaultEffect(t *testing.T) {
	detector := NewEffectConflictDetector()
	rm := defaultrolemanager.NewRoleManagerImpl(10)
	
	// Create a model
	m := model.Model{}
	m.AddDef("p", "p", "sub, obj, act")
	m.AddDef("g", "g", "_, _")
	
	// alice has default effect (allow)
	_ = m.AddPolicy("p", "p", []string{"alice", "data1", "read"})
	// admin is denied
	_ = m.AddPolicy("p", "p", []string{"admin", "data1", "read", "deny"})
	
	// alice has role admin
	_ = m.AddPolicy("g", "g", []string{"alice", "admin"})
	
	err := detector.CheckModel(m, rm)
	if err == nil {
		t.Error("Expected conflict detection error with default effect, but got nil")
	} else {
		errMsg := err.Error()
		if !strings.Contains(errMsg, "effect conflict detected") {
			t.Errorf("Expected error message to contain 'effect conflict detected', got: %s", errMsg)
		}
	}
}

func TestEffectConflictDetector_DifferentActions(t *testing.T) {
	detector := NewEffectConflictDetector()
	rm := defaultrolemanager.NewRoleManagerImpl(10)
	
	// Create a model
	m := model.Model{}
	m.AddDef("p", "p", "sub, obj, act, eft")
	m.AddDef("g", "g", "_, _")
	
	// alice is allowed to read data1
	_ = m.AddPolicy("p", "p", []string{"alice", "data1", "read", "allow"})
	// admin is denied to write data1 (different action)
	_ = m.AddPolicy("p", "p", []string{"admin", "data1", "write", "deny"})
	
	// alice has role admin
	_ = m.AddPolicy("g", "g", []string{"alice", "admin"})
	
	err := detector.CheckModel(m, rm)
	if err != nil {
		t.Errorf("Expected no conflict for different actions, but got error: %v", err)
	}
}

func TestEffectConflictDetector_DifferentObjects(t *testing.T) {
	detector := NewEffectConflictDetector()
	rm := defaultrolemanager.NewRoleManagerImpl(10)
	
	// Create a model
	m := model.Model{}
	m.AddDef("p", "p", "sub, obj, act, eft")
	m.AddDef("g", "g", "_, _")
	
	// alice is allowed to read data1
	_ = m.AddPolicy("p", "p", []string{"alice", "data1", "read", "allow"})
	// admin is denied to read data2 (different object)
	_ = m.AddPolicy("p", "p", []string{"admin", "data2", "read", "deny"})
	
	// alice has role admin
	_ = m.AddPolicy("g", "g", []string{"alice", "admin"})
	
	err := detector.CheckModel(m, rm)
	if err != nil {
		t.Errorf("Expected no conflict for different objects, but got error: %v", err)
	}
}
