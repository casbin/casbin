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

package casbin

import (
	"testing"

	"github.com/casbin/casbin/v2/persist/file-adapter"
)

func TestPathError(t *testing.T) {
	_, err := NewEnforcer("hope_this_path_wont_exist", "")
	if err == nil {
		t.Errorf("Should be error here.")
	} else {
		t.Log("Test on error: ")
		t.Log(err.Error())
	}
}

func TestEnforcerParamError(t *testing.T) {
	_, err := NewEnforcer(1, 2, 3)
	if err == nil {
		t.Errorf("Should not be error here.")
	} else {
		t.Log("Test on error: ")
		t.Log(err.Error())
	}

	_, err2 := NewEnforcer(1, "2")
	if err2 == nil {
		t.Errorf("Should not be error here.")
	} else {
		t.Log("Test on error: ")
		t.Log(err2.Error())
	}
}

func TestModelError(t *testing.T) {
	_, err := NewEnforcer("examples/error/error_model.conf", "examples/error/error_policy.csv")
	if err == nil {
		t.Errorf("Should be error here.")
	} else {
		t.Log("Test on error: ")
		t.Log(err.Error())
	}
}

//func TestPolicyError(t *testing.T) {
//	_, err := NewEnforcer("examples/basic_model.conf", "examples/error/error_policy.csv")
//	if err == nil {
//		t.Errorf("Should be error here.")
//	} else {
//		t.Log("Test on error: ")
//		t.Log(err.Error())
//	}
//}

func TestEnforceError(t *testing.T) {
	e, _ := NewEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")

	_, err := e.Enforce("wrong", "wrong")
	if err == nil {
		t.Errorf("Should be error here.")
	} else {
		t.Log("Test on error: ")
		t.Log(err.Error())
	}
}

func TestNoError(t *testing.T) {
	e, _ := NewEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")

	err := e.LoadModel()
	if err != nil {
		t.Errorf("Should be no error here.")
		t.Log("Unexpected error: ")
		t.Log(err.Error())
	}

	err = e.LoadPolicy()
	if err != nil {
		t.Errorf("Should be no error here.")
		t.Log("Unexpected error: ")
		t.Log(err.Error())
	}

	err = e.SavePolicy()
	if err != nil {
		t.Errorf("Should be no error here.")
		t.Log("Unexpected error: ")
		t.Log(err.Error())
	}
}

func TestModelNoError(t *testing.T) {
	e, _ := NewEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")

	e.modelPath = "hope_this_path_wont_exist"
	err := e.LoadModel()

	if err == nil {
		t.Errorf("Should be error here.")
	} else {
		t.Log("Test on error: ")
		t.Log(err.Error())
	}
}

func TestMockAdapterErrors(t *testing.T) {
	adapter := fileadapter.NewAdapterMock("examples/rbac_with_domains_policy.csv")
	adapter.SetMockErr("mock error")

	e, _ := NewEnforcer("examples/rbac_with_domains_model.conf", adapter)

	_, err := e.AddPolicy("admin", "domain3", "data1", "read")

	if err == nil {
		t.Errorf("Should be an error here.")
	} else {
		t.Log("Test on error: ")
		t.Log(err.Error())
	}

	rules := [][]string {
			{"admin", "domain4", "data1", "read"},
	}
	_, err = e.AddPolicies(rules)

	if err == nil {
		t.Errorf("Should be an error here.")
	} else {
		t.Log("Test on error: ")
		t.Log(err.Error())
	}

	_, err2 := e.RemoveFilteredPolicy(1, "domain1", "data1")

	if err2 == nil {
		t.Errorf("Should be an error here.")
	} else {
		t.Log("Test on error: ")
		t.Log(err2.Error())
	}

	_, err3 := e.RemovePolicy("admin", "domain2", "data2", "read")

	if err3 == nil {
		t.Errorf("Should be an error here.")
	} else {
		t.Log("Test on error: ")
		t.Log(err3.Error())
	}

	rules = [][]string {
		{"admin", "domain4", "data1", "read"},
	}

	_, err = e.RemovePolicies(rules)

	if err == nil {
		t.Errorf("Should be an error here.")
	} else {
		t.Log("Test on error: ")
		t.Log(err.Error())
	}
	_, err4 := e.AddGroupingPolicy("bob", "admin2")

	if err4 == nil {
		t.Errorf("Should be an error here.")
	} else {
		t.Log("Test on error: ")
		t.Log(err4.Error())
	}

	_, err5 := e.AddNamedGroupingPolicy("g", []string{"eve", "admin2", "domain1"})

	if err5 == nil {
		t.Errorf("Should be an error here.")
	} else {
		t.Log("Test on error: ")
		t.Log(err5.Error())
	}

	_, err6 := e.AddNamedPolicy("p", []string{"admin2", "domain2", "data2", "write"})

	if err6 == nil {
		t.Errorf("Should be an error here.")
	} else {
		t.Log("Test on error: ")
		t.Log(err6.Error())
	}

	_, err7 := e.RemoveGroupingPolicy("bob", "admin2")

	if err7 == nil {
		t.Errorf("Should be an error here.")
	} else {
		t.Log("Test on error: ")
		t.Log(err7.Error())
	}

	_, err8 := e.RemoveFilteredGroupingPolicy(0, "bob")

	if err8 == nil {
		t.Errorf("Should be an error here.")
	} else {
		t.Log("Test on error: ")
		t.Log(err8.Error())
	}

	_, err9 := e.RemoveNamedGroupingPolicy("g", []string{"alice", "admin", "domain1"})

	if err9 == nil {
		t.Errorf("Should be an error here.")
	} else {
		t.Log("Test on error: ")
		t.Log(err9.Error())
	}

	_, err10 := e.RemoveFilteredNamedGroupingPolicy("g", 0, "eve")

	if err10 == nil {
		t.Errorf("Should be an error here.")
	} else {
		t.Log("Test on error: ")
		t.Log(err10.Error())
	}
}
