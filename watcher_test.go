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

import "testing"

type SampleWatcher struct {
}

func (w SampleWatcher) Close() {
}

func (w SampleWatcher) SetUpdateCallback(func(string)) error {
	return nil
}

func (w SampleWatcher) Update() error {
	return nil
}

func TestSetWatcher(t *testing.T) {
	e, err := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	if err != nil {
		t.Fatal(err)
	}
	sampleWatcher := SampleWatcher{}
	err = e.SetWatcher(sampleWatcher)
	if err != nil {
		t.Fatal(err)
	}
	err = e.SavePolicy() //calls watcher.Update()
	if err != nil {
		t.Fatal(err)
	}
}
