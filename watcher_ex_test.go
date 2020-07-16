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
	"testing"

	"github.com/casbin/casbin/v3/model"
)

type SampleWatcherEx struct {
	SampleWatcher
}

func (w SampleWatcherEx) UpdateForAddPolicy(params ...string) error {
	return nil
}
func (w SampleWatcherEx) UpdateForRemovePolicy(params ...string) error {
	return nil
}

func (w SampleWatcherEx) UpdateForRemoveFilteredPolicy(fieldIndex int, fieldValues ...string) error {
	return nil
}

func (w SampleWatcherEx) UpdateForSavePolicy(model model.Model) error {
	return nil
}

func TestSetWatcherEx(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")

	sampleWatcherEx := SampleWatcherEx{}
	err := e.SetWatcher(sampleWatcherEx)
	if err != nil {
		t.Fatal(err)
	}

	_ = e.SavePolicy()                              // calls watcherEx.UpdateForSavePolicy()
	_, _ = e.AddPolicy("admin", "data1", "read")    // calls watcherEx.UpdateForAddPolicy()
	_, _ = e.RemovePolicy("admin", "data1", "read") // calls watcherEx.UpdateForRemovePolicy()
	_, _ = e.RemoveFilteredPolicy(1, "data1")       // calls watcherEx.UpdateForRemoveFilteredPolicy()
}
