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
	// EventEnforce represents an enforcement event.
	EventEnforce EventType = "enforce"
	// EventPolicyAdd represents a policy addition event.
	EventPolicyAdd EventType = "policy_add"
	// EventPolicyRemove represents a policy removal event.
	EventPolicyRemove EventType = "policy_remove"
	// EventPolicyUpdate represents a policy update event.
	EventPolicyUpdate EventType = "policy_update"
	// EventPolicyLoad represents a policy load event.
	EventPolicyLoad EventType = "policy_load"
	// EventPolicySave represents a policy save event.
	EventPolicySave EventType = "policy_save"
	// EventModelLoad represents a model load event.
	EventModelLoad EventType = "model_load"
	// EventRoleAdd represents a role addition event.
	EventRoleAdd EventType = "role_add"
	// EventRoleRemove represents a role removal event.
	EventRoleRemove EventType = "role_remove"
)

// Handle contains context information for an ongoing logging event.
type Handle struct {
	// StartTime is the timestamp when the event started.
	StartTime time.Time
	// Store is a map for storing arbitrary key-value data during event processing.
	Store map[string]interface{}
}

// LogEntry represents a complete log entry for a Casbin event.
type LogEntry struct {
	// EventType is the type of the event being logged.
	EventType EventType
	// Timestamp is when the event occurred.
	Timestamp time.Time
	// Duration is how long the event took to complete.
	Duration time.Duration

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
	// Matched indicates which policy rule matched the request.
	Matched string

	// Policy operation parameters.
	// Operation describes the type of policy operation (add, remove, update).
	Operation string
	// Rules contains the policy rules involved in the operation.
	Rules [][]string
	// RuleCount is the number of rules affected by the operation.
	RuleCount int

	// Error contains any error that occurred during the event.
	Error error
	// Attributes stores additional arbitrary metadata for the event.
	Attributes map[string]interface{}
}
