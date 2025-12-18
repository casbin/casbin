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

import "time"

// Handle is passed from OnBeforeEvent to OnAfterEvent.
// Logger implementations can store custom data in the Store field.
type Handle struct {
	// StartTime records when the event started.
	StartTime time.Time

	// Store allows logger implementations to attach custom data.
	// e.g., OpenTelemetry can store Span, context, etc.
	Store map[string]interface{}
}

// NewHandle creates a new Handle with initialized fields.
func NewHandle() *Handle {
	return &Handle{
		StartTime: time.Now(),
		Store:     make(map[string]interface{}),
	}
}
