// Copyright 2025 The casbin Authors. All Rights Reserved.
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

package util

import (
	"testing"
)

func TestCheckSecurityLevel(t *testing.T) {
	// Each test case checks if the Bell-LaPadula model enforces the correct access control
	tests := []struct {
		action      string
		levels      []float64
		expectAllow bool
		desc        string
	}{
		// Read: subject can only read objects at the same or lower security level (no read-up)
		{"read", []float64{3, 1}, true, "Top secret can read public document"},
		{"read", []float64{3, 2}, true, "Top secret can read secret document"},
		{"read", []float64{3, 3}, true, "Top secret can read top secret document"},
		{"read", []float64{2, 1}, true, "Secret can read public document"},
		{"read", []float64{2, 2}, true, "Secret can read secret document"},
		{"read", []float64{2, 3}, false, "Secret cannot read top secret document"},
		{"read", []float64{1, 1}, true, "Public can read public document"},
		{"read", []float64{1, 2}, false, "Public cannot read secret document"},
		{"read", []float64{1, 3}, false, "Public cannot read top secret document"},

		// Write: subject cannot write to objects at a lower security level (no write-down)
		{"write", []float64{3, 1}, false, "Top secret cannot write to public document"},
		{"write", []float64{3, 2}, false, "Top secret cannot write to secret document"},
		{"write", []float64{3, 3}, true, "Top secret can write to top secret document"},
		{"write", []float64{2, 1}, false, "Secret cannot write to public document"},
		{"write", []float64{2, 2}, true, "Secret can write to secret document"},
		{"write", []float64{2, 3}, true, "Secret can write to top secret document"},
		{"write", []float64{1, 1}, true, "Public can write to public document"},
		{"write", []float64{1, 2}, true, "Public can write to secret document"},
		{"write", []float64{1, 3}, true, "Public can write to top secret document"},

		// Invalid actions should always be denied
		{"delete", []float64{3, 1}, false, "Invalid action should be denied"},
		{"", []float64{3, 1}, false, "Empty action should be denied"},
		// Edge cases for lowest security level
		{"read", []float64{0, 0}, true, "Level 0 can read level 0"},
		{"write", []float64{0, 0}, true, "Level 0 can write to level 0"},
		{"read", []float64{0, 1}, false, "Level 0 cannot read level 1"},
		{"write", []float64{1, 0}, false, "Level 1 cannot write to level 0"},
	}

	for _, tc := range tests {
		allowed := CheckSecurityLevel(tc.action, tc.levels[0], tc.levels[1])
		if allowed != tc.expectAllow {
			t.Errorf("%s: action=%s, levels=%v, expect=%v, got=%v", tc.desc, tc.action, tc.levels, tc.expectAllow, allowed)
		}
	}
}
