// Copyright 2020 The casbin Authors. All Rights Reserved.
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
	"io/ioutil"
	"regexp"
	"strings"
	"testing"
)

func TestCasbinJsGetPermissionForUser(t *testing.T) {
	e, err := NewSyncedEnforcer("examples/rbac_model.conf", "examples/rbac_with_hierarchy_policy.csv")
	if err != nil {
		panic(err)
	}
	receivedString, err := CasbinJsGetPermissionForUser(e, "alice") // make sure CasbinJsGetPermissionForUser can be used with a SyncedEnforcer.
	if err != nil {
		t.Errorf("Test error: %s", err)
	}
	received := map[string]interface{}{}
	err = json.Unmarshal([]byte(receivedString), &received)
	if err != nil {
		t.Errorf("Test error: %s", err)
	}
	expectedModel, err := ioutil.ReadFile("examples/rbac_model.conf")
	if err != nil {
		t.Errorf("Test error: %s", err)
	}
	expectedModelStr := regexp.MustCompile("\n+").ReplaceAllString(string(expectedModel), "\n")
	if strings.TrimSpace(received["m"].(string)) != expectedModelStr {
		t.Errorf("%s supposed to be %s", strings.TrimSpace(received["m"].(string)), expectedModelStr)
	}

	expectedPolicies, err := ioutil.ReadFile("examples/rbac_with_hierarchy_policy.csv")
	if err != nil {
		t.Errorf("Test error: %s", err)
	}
	expectedPoliciesItem := regexp.MustCompile(",|\n").Split(string(expectedPolicies), -1)
	i := 0
	for _, sArr := range received["p"].([]interface{}) {
		for _, s := range sArr.([]interface{}) {
			if strings.TrimSpace(s.(string)) != strings.TrimSpace(expectedPoliciesItem[i]) {
				t.Errorf("%s supposed to be %s", strings.TrimSpace(s.(string)), strings.TrimSpace(expectedPoliciesItem[i]))
			}
			i++
		}
	}
}
