// Copyright 2021 The casbin Authors. All Rights Reserved.
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
	"encoding/json"
	"testing"
)

func contains(arr []string, target string) bool {
	for _, item := range arr {
		if item == target {
			return true
		}
	}
	return false
}

func TestCasbinJsGetPermissionForUserOld(t *testing.T) {
	e, err := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	if err != nil {
		panic(err)
	}
	target_str, _ := CasbinJsGetPermissionForUserOld(e, "alice")
	t.Log("GetPermissionForUser Alice", string(target_str))
	alice_target := make(map[string][]string)
	err = json.Unmarshal(target_str, &alice_target)
	if err != nil {
		t.Errorf("Test error: %s", err)
	}
	perm, ok := alice_target["read"]
	if !ok {
		t.Errorf("Test error: Alice doesn't have read permission")
	}
	if !contains(perm, "data1") {
		t.Errorf("Test error: Alice cannot read data1")
	}
	if !contains(perm, "data2") {
		t.Errorf("Test error: Alice cannot read data2")
	}
	perm, ok = alice_target["write"]
	if !ok {
		t.Errorf("Test error: Alice doesn't have write permission")
	}
	if contains(perm, "data1") {
		t.Errorf("Test error: Alice can write data1")
	}
	if !contains(perm, "data2") {
		t.Errorf("Test error: Alice cannot write data2")
	}

	target_str, _ = CasbinJsGetPermissionForUserOld(e, "bob")
	t.Log("GetPermissionForUser Bob", string(target_str))
	bob_target := make(map[string][]string)
	err = json.Unmarshal(target_str, &bob_target)
	if err != nil {
		t.Errorf("Test error: %s", err)
	}
	_, ok = bob_target["read"]
	if ok {
		t.Errorf("Test error: Bob has read permission")
	}
	perm, ok = bob_target["write"]
	if !ok {
		t.Errorf("Test error: Bob doesn't have permission")
	}
	if !contains(perm, "data2") {
		t.Errorf("Test error: Bob cannot write data2")
	}
	if contains(perm, "data1") {
		t.Errorf("Test error: Bob can write data1")
	}
	if contains(perm, "data_not_exist") {
		t.Errorf("Test error: Bob can access a non-existing data")
	}

	_, ok = bob_target["rm_rf"]
	if ok {
		t.Errorf("Someone can have a non-existing action (rm -rf)")
	}
}
