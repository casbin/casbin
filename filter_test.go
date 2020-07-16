// Copyright 2018 The casbin Authors. All Rights Reserved.
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

	"github.com/casbin/casbin/v3/persist/file-adapter"
)

func TestInitFilteredAdapter(t *testing.T) {
	e, _ := NewEnforcer()

	adapter := fileadapter.NewFilteredAdapter("examples/rbac_with_domains_policy.csv")
	_ = e.InitWithAdapter("examples/rbac_with_domains_model.conf", adapter)

	// policy should not be loaded yet
	testHasPolicy(t, e, []string{"admin", "domain1", "data1", "read"}, false)
}

func TestLoadFilteredPolicy(t *testing.T) {
	e, _ := NewEnforcer()

	adapter := fileadapter.NewFilteredAdapter("examples/rbac_with_domains_policy.csv")
	_ = e.InitWithAdapter("examples/rbac_with_domains_model.conf", adapter)
	if err := e.LoadPolicy(); err != nil {
		t.Errorf("unexpected error in LoadPolicy: %v", err)
	}

	// validate initial conditions
	testHasPolicy(t, e, []string{"admin", "domain1", "data1", "read"}, true)
	testHasPolicy(t, e, []string{"admin", "domain2", "data2", "read"}, true)

	if err := e.LoadFilteredPolicy(&fileadapter.Filter{
		P: []string{"", "domain1"},
		G: []string{"", "", "domain1"},
	}); err != nil {
		t.Errorf("unexpected error in LoadFilteredPolicy: %v", err)
	}
	if !e.IsFiltered() {
		t.Errorf("adapter did not set the filtered flag correctly")
	}

	// only policies for domain1 should be loaded
	testHasPolicy(t, e, []string{"admin", "domain1", "data1", "read"}, true)
	testHasPolicy(t, e, []string{"admin", "domain2", "data2", "read"}, false)

	if err := e.SavePolicy(); err == nil {
		t.Errorf("enforcer did not prevent saving filtered policy")
	}
	if err := e.GetAdapter().SavePolicy(e.GetModel()); err == nil {
		t.Errorf("adapter did not prevent saving filtered policy")
	}
}

func TestFilteredPolicyInvalidFilter(t *testing.T) {
	e, _ := NewEnforcer()

	adapter := fileadapter.NewFilteredAdapter("examples/rbac_with_domains_policy.csv")
	_ = e.InitWithAdapter("examples/rbac_with_domains_model.conf", adapter)

	if err := e.LoadFilteredPolicy([]string{"", "domain1"}); err == nil {
		t.Errorf("expected error in LoadFilteredPolicy, but got nil")
	}
}

func TestFilteredPolicyEmptyFilter(t *testing.T) {
	e, _ := NewEnforcer()

	adapter := fileadapter.NewFilteredAdapter("examples/rbac_with_domains_policy.csv")
	_ = e.InitWithAdapter("examples/rbac_with_domains_model.conf", adapter)

	if err := e.LoadFilteredPolicy(nil); err != nil {
		t.Errorf("unexpected error in LoadFilteredPolicy: %v", err)
	}
	if e.IsFiltered() {
		t.Errorf("adapter did not reset the filtered flag correctly")
	}
	if err := e.SavePolicy(); err != nil {
		t.Errorf("unexpected error in SavePolicy: %v", err)
	}
}

func TestUnsupportedFilteredPolicy(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_with_domains_model.conf", "examples/rbac_with_domains_policy.csv")

	err := e.LoadFilteredPolicy(&fileadapter.Filter{
		P: []string{"", "domain1"},
		G: []string{"", "", "domain1"},
	})
	if err == nil {
		t.Errorf("encorcer should have reported incompatibility error")
	}
}

func TestFilteredAdapterEmptyFilepath(t *testing.T) {
	e, _ := NewEnforcer()

	adapter := fileadapter.NewFilteredAdapter("")
	_ = e.InitWithAdapter("examples/rbac_with_domains_model.conf", adapter)

	if err := e.LoadFilteredPolicy(nil); err != nil {
		t.Errorf("unexpected error in LoadFilteredPolicy: %v", err)
	}
}

func TestFilteredAdapterInvalidFilepath(t *testing.T) {
	e, _ := NewEnforcer()

	adapter := fileadapter.NewFilteredAdapter("examples/does_not_exist_policy.csv")
	_ = e.InitWithAdapter("examples/rbac_with_domains_model.conf", adapter)

	if err := e.LoadFilteredPolicy(nil); err == nil {
		t.Errorf("expected error in LoadFilteredPolicy, but got nil")
	}
}
