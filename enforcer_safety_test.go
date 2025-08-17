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

package casbin

import (
	"testing"
	"time"
)

func TestSafeEnforcer(t *testing.T) {
	e, err := NewSafeEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")
	if err != nil {
		t.Errorf("NewSafeEnforcer failed: %v", err)
		return
	}
	
	// Test SafeEnforce
	result, err := e.SafeEnforce("alice", "data1", "read")
	if err != nil {
		t.Errorf("SafeEnforce failed: %v", err)
	}
	if !result {
		t.Error("SafeEnforce result should be true")
	}
	
	// Test EnforceWithValidation
	result, err = e.EnforceWithValidation("alice", "data1", "read")
	if err != nil {
		t.Errorf("EnforceWithValidation failed: %v", err)
	}
	if !result {
		t.Error("EnforceWithValidation result should be true")
	}
	
	// Test invalid parameters
	_, err = e.EnforceWithValidation()
	if err == nil {
		t.Error("EnforceWithValidation should return error with empty parameters")
	}
	
	// Test RecoverableEnforce
	result, err = e.RecoverableEnforce("alice", "data1", "read")
	if err != nil {
		t.Errorf("RecoverableEnforce failed: %v", err)
	}
	if !result {
		t.Error("RecoverableEnforce result should be true")
	}
	
	// Test HealthCheck
	err = e.HealthCheck()
	if err != nil {
		t.Errorf("HealthCheck failed: %v", err)
	}
}

func TestSafeEnforcerTimeout(t *testing.T) {
	e, err := NewSafeEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")
	if err != nil {
		t.Errorf("NewSafeEnforcer failed: %v", err)
		return
	}
	
	// Set a very short timeout to test timeout functionality
	e.SetDefaultTimeout(1 * time.Nanosecond)
	
	// Create a slow operation that will likely timeout
	slowOperation := func() {
		time.Sleep(100 * time.Millisecond)
	}
	
	go slowOperation()
	
	// This should timeout
	_, err = e.LoadPolicyWithTimeout(1 * time.Nanosecond)
	if err == nil || err.Error() != "load policy operation timed out" {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

func TestSafeEnforcerWithInvalidModel(t *testing.T) {
	// Test with non-existent model file
	_, err := NewSafeEnforcer("non_existent_model.conf", "examples/basic_policy.csv")
	if err == nil {
		t.Error("NewSafeEnforcer should return error with non-existent model file")
	}
}