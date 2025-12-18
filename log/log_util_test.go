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

package log

import (
	"testing"
)

func TestLogWithTestLogger(t *testing.T) {
	tl := NewTestLogger()
	SetLogger(tl)

	// Test Enable/IsEnabled
	tl.EnableLog(true)
	if !tl.IsEnabled() {
		t.Error("Logger should be enabled")
	}

	tl.EnableLog(false)
	if tl.IsEnabled() {
		t.Error("Logger should be disabled")
	}

	tl.EnableLog(true)

	// Test LogPolicy (no-op, just ensure no panic)
	policy := map[string][][]string{}
	LogPolicy(policy)

	// Test LogModel (no-op)
	var model [][]string
	LogModel(model)

	// Test LogEnforce (no-op)
	LogEnforce("my_matcher", []interface{}{"bob"}, true, nil)

	// Test LogRole (no-op)
	LogRole([]string{})

	// Test Subscribe
	if Subscribe() != nil {
		t.Error("Default subscribe should be nil")
	}

	tl.SetSubscribe([]EventType{EventEnforce})
	events := Subscribe()
	if len(events) != 1 || events[0] != EventEnforce {
		t.Error("Subscribe should return EventEnforce")
	}

	// Test OnBeforeEvent/OnAfterEvent
	entry := NewLogEntry(EventEnforce)
	entry.Subject = "alice"

	handle := OnBeforeEvent(entry)
	if handle == nil {
		t.Error("OnBeforeEvent should return a Handle")
	}
	if tl.LastEntry != entry {
		t.Error("TestLogger should capture the entry")
	}

	OnAfterEvent(handle, entry)
	if tl.LastHandle != handle {
		t.Error("TestLogger should capture the handle")
	}
}
