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

// LogEntry contains all information about an event.
type LogEntry struct {
	// Event info
	Type      EventType
	Timestamp time.Time
	Duration  time.Duration // Filled in OnAfterEvent

	// Enforce related
	Request []interface{}
	Subject string
	Object  string
	Action  string
	Domain  string
	Allowed bool
	Matched [][]string

	// Policy/Role related
	Operation string
	Rules     [][]string
	RuleCount int

	// Error info
	Error error

	// Custom attributes (can store context.Context, trace IDs, etc.)
	Attributes map[string]interface{}
}

// NewLogEntry creates a new LogEntry with initialized fields.
func NewLogEntry(eventType EventType) *LogEntry {
	return &LogEntry{
		Type:       eventType,
		Timestamp:  time.Now(),
		Attributes: make(map[string]interface{}),
	}
}
