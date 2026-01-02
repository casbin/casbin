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

package log

import "time"

// EventType represents the type of logging event.
type EventType string

// Event type constants.
const (
	EventEnforce      EventType = "enforce"
	EventAddPolicy    EventType = "addPolicy"
	EventRemovePolicy EventType = "removePolicy"
	EventLoadPolicy   EventType = "loadPolicy"
	EventSavePolicy   EventType = "savePolicy"
)

// LogEntry represents a complete log entry for a Casbin event.
type LogEntry struct {
	IsActive bool
	// EventType is the type of the event being logged.
	EventType EventType

	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration

	// Enforce parameters.
	// Subject is the user or entity requesting access.
	Subject string
	// Object is the resource being accessed.
	Object string
	// Action is the operation being performed.
	Action string
	// Domain is the domain/tenant for multi-tenant scenarios.
	Domain string
	// Allowed indicates whether the enforcement request was allowed.
	Allowed bool

	// Rules contains the policy rules involved in the operation.
	Rules [][]string
	// RuleCount is the number of rules affected by the operation.
	RuleCount int

	// Error contains any error that occurred during the event.
	Error error
}
