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
	"fmt"
	"testing"

	"github.com/casbin/casbin/persist/file-adapter"
)

func TestPathError(t *testing.T) {
	_, err := NewEnforcerSafe("hope_this_path_wont_exist", "")
	if err == nil {
		t.Errorf("Should be error here.")
	} else {
		fmt.Print("Test on error: ")
		fmt.Print(err.Error())
	}
}

func TestEnforcerParamError(t *testing.T) {
	_, err := NewEnforcerSafe(1, 2, 3)
	if err == nil {
		t.Errorf("Should not be error here.")
	} else {
		fmt.Print("Test on error: ")
		fmt.Print(err.Error())
	}

	_, err2 := NewEnforcerSafe(1, "2")
	if err2 == nil {
		t.Errorf("Should not be error here.")
	} else {
		fmt.Print("Test on error: ")
		fmt.Print(err2.Error())
	}
}

func TestModelError(t *testing.T) {
	_, err := NewEnforcerSafe("examples/error/error_model.conf", "examples/error/error_policy.csv")
	if err == nil {
		t.Errorf("Should be error here.")
	} else {
		fmt.Print("Test on error: ")
		fmt.Print(err.Error())
	}
}

func TestPolicyError(t *testing.T) {
	_, err := NewEnforcerSafe("examples/basic_model.conf", "examples/error/error_policy.csv")
	if err == nil {
		t.Errorf("Should be error here.")
	} else {
		fmt.Print("Test on error: ")
		fmt.Print(err.Error())
	}
}

func TestEnforceError(t *testing.T) {
	e := NewEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")

	_, err := e.EnforceSafe("wrong", "wrong")
	if err == nil {
		t.Errorf("Should be error here.")
	} else {
		fmt.Print("Test on error: ")
		fmt.Print(err.Error())
	}
}

func TestNoError(t *testing.T) {
	e := NewEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")

	err := e.LoadModelSafe()
	if err != nil {
		t.Errorf("Should be no error here.")
		fmt.Print("Unexpected error: ")
		fmt.Print(err.Error())
	}

	err = e.LoadPolicy()
	if err != nil {
		t.Errorf("Should be no error here.")
		fmt.Print("Unexpected error: ")
		fmt.Print(err.Error())
	}

	err = e.SavePolicy()
	if err != nil {
		t.Errorf("Should be no error here.")
		fmt.Print("Unexpected error: ")
		fmt.Print(err.Error())
	}
}

func TestModelNoError(t *testing.T) {
	e := NewEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")

	e.modelPath = "hope_this_path_wont_exist"
	err := e.LoadModelSafe()

	if err == nil {
		t.Errorf("Should be error here.")
	} else {
		fmt.Print("Test on error: ")
		fmt.Print(err.Error())
	}
}

func TestMockAdapterErrors(t *testing.T) {
	adapter := fileadapter.NewAdapterMock("examples/rbac_with_domains_policy.csv")
	adapter.SetMockErr("mock error")

	e, _ := NewEnforcerSafe("examples/rbac_with_domains_model.conf", adapter)

	_, err := e.AddPolicySafe("admin", "domain3", "data1", "read")

	if err == nil {
		t.Errorf("Should be an error here.")
	} else {
		fmt.Print("Test on error: ")
		fmt.Print(err.Error())
	}

	_, err2 := e.RemoveFilteredPolicySafe(1, "domain1", "data1")

	if err2 == nil {
		t.Errorf("Should be an error here.")
	} else {
		fmt.Print("Test on error: ")
		fmt.Print(err2.Error())
	}

	_, err3 := e.RemovePolicySafe("admin", "domain2", "data2", "read")

	if err3 == nil {
		t.Errorf("Should be an error here.")
	} else {
		fmt.Print("Test on error: ")
		fmt.Print(err3.Error())
	}
}
