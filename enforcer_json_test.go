// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package casbin

import (
	"strings"
	"testing"

	"github.com/casbin/casbin/v3/model"
)

func TestInvalidJsonRequest(t *testing.T) {
	modelText := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub.Name == " "
`

	m, err := model.NewModelFromString(modelText)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	e, err := NewEnforcer(m)
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}
	e.EnableAcceptJsonRequest(true)

	// Test with invalid JSON (contains \x escape sequence which is not valid in JSON)
	invalidJSON := `{"Name": "\x20"}`
	_, err = e.Enforce(invalidJSON, "obj", "read")
	if err == nil {
		t.Fatalf("Expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse JSON parameter") {
		t.Fatalf("Expected error message to contain 'failed to parse JSON parameter', got: %v", err)
	}

	// Test with valid JSON - should work
	validJSON := `{"Name": " "}`
	res, err := e.Enforce(validJSON, "obj", "read")
	if err != nil {
		t.Fatalf("Valid JSON should not return error: %v", err)
	}
	if !res {
		t.Fatalf("Expected true for valid JSON with matching Name")
	}

	// Test with plain string (doesn't start with { or [) - should not try to parse as JSON
	plainString := "alice"
	_, err = e.Enforce(plainString, "obj", "read")
	// This will fail because plainString is not a struct with Name field,
	// but it shouldn't fail with JSON parsing error
	if err != nil && strings.Contains(err.Error(), "failed to parse JSON parameter") {
		t.Fatalf("Plain string should not trigger JSON parsing error: %v", err)
	}
}
