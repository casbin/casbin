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

package persist_test

import (
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
)

func TestPersist(t *testing.T) {
	//No tests yet
}

func testRuleCount(t *testing.T, model model.Model, expected int, sec string, ptype string, tag string) {
	t.Helper()

	ruleCount := len(model[sec][ptype].Policy)
	if ruleCount != expected {
		t.Errorf("[%s] rule count: %d, expected %d", tag, ruleCount, expected)
	}
}

func TestDuplicateRuleInAdapter(t *testing.T) {
	e, _ := casbin.NewEnforcer("../examples/basic_model.conf")

	_, _ = e.AddPolicy("alice", "data1", "read")
	_, _ = e.AddPolicy("alice", "data1", "read")

	testRuleCount(t, e.GetModel(), 1, "p", "p", "AddPolicy")

	e.ClearPolicy()

	//simulate adapter.LoadPolicy with duplicate rules
	_ = persist.LoadPolicyArray([]string{"p", "alice", "data1", "read"}, e.GetModel())
	_ = persist.LoadPolicyArray([]string{"p", "alice", "data1", "read"}, e.GetModel())

	testRuleCount(t, e.GetModel(), 1, "p", "p", "LoadPolicyArray")
}
