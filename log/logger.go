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

package log

import "time"

//go:generate mockgen -destination=./mocks/mock_logger.go -package=mocks github.com/casbin/casbin/v3/log Logger

// EventType represents the type of event being logged.
type EventType string

const (
	EventEnforce      EventType = "enforce"
	EventPolicyAdd    EventType = "policy.add"
	EventPolicyRemove EventType = "policy.remove"
	EventPolicyUpdate EventType = "policy.update"
	EventPolicyLoad   EventType = "policy.load"
	EventPolicySave   EventType = "policy.save"
	EventModelLoad    EventType = "model.load"
	EventRoleAdd      EventType = "role.add"
	EventRoleRemove   EventType = "role.remove"
)

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

// Logger is the logging interface implementation.
type Logger interface {
	// EnableLog controls whether print the message.
	EnableLog(bool)

	// IsEnabled returns if logger is enabled.
	IsEnabled() bool

	// LogModel log info related to model.
	LogModel(model [][]string)

	// LogEnforce log info related to enforce.
	LogEnforce(matcher string, request []interface{}, result bool, explains [][]string)

	// LogRole log info related to role.
	LogRole(roles []string)

	// LogPolicy log info related to policy.
	LogPolicy(policy map[string][][]string)

	// LogError log info relate to error
	LogError(err error, msg ...string)

	// Subscribe returns the list of event types this logger is interested in.
	// Return nil or empty slice to subscribe to all events.
	// Return specific event types to filter events.
	Subscribe() []EventType

	// OnBeforeEvent is called before an event occurs.
	// Returns a Handle that will be passed to OnAfterEvent.
	OnBeforeEvent(entry *LogEntry) *Handle

	// OnAfterEvent is called after an event completes.
	// The Handle from OnBeforeEvent is passed back along with the updated entry.
	OnAfterEvent(handle *Handle, entry *LogEntry)
}
