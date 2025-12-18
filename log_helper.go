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
	"time"

	"github.com/casbin/casbin/v3/log"
)

// LogEnforceEvent records an enforce event and returns a finish function.
// Usage:
//
//	finish, setResult := e.LogEnforceEvent(rvals)
//	defer finish()
//	// ... do enforce logic ...
//	setResult(allowed, matched, err)
func (e *Enforcer) LogEnforceEvent(rvals []interface{}) (finish func(), setResult func(bool, [][]string, error)) {
	if !e.shouldLog(log.EventEnforce) {
		return func() {}, func(bool, [][]string, error) {}
	}

	entry := log.NewLogEntry(log.EventEnforce)
	entry.Request = rvals

	// Parse request parameters
	if len(rvals) >= 1 {
		if s, ok := rvals[0].(string); ok {
			entry.Subject = s
		}
	}
	if len(rvals) >= 2 {
		if s, ok := rvals[1].(string); ok {
			entry.Object = s
		}
	}
	if len(rvals) >= 3 {
		if s, ok := rvals[2].(string); ok {
			entry.Action = s
		}
	}
	if len(rvals) >= 4 {
		if s, ok := rvals[3].(string); ok {
			entry.Domain = s
		}
	}

	handle := e.logger.OnBeforeEvent(entry)

	setResult = func(allowed bool, matched [][]string, err error) {
		entry.Allowed = allowed
		entry.Matched = matched
		entry.Error = err
	}

	finish = func() {
		entry.Duration = time.Since(entry.Timestamp)
		e.logger.OnAfterEvent(handle, entry)
	}

	return finish, setResult
}

// LogPolicyEvent records a policy-related event.
// Usage:
//
//	var err error
//	defer e.LogPolicyEvent(log.EventPolicyAdd, "add", rules, &err)()
//	// ... do policy operation ...
//	err = actualError
func (e *Enforcer) LogPolicyEvent(
	eventType log.EventType,
	operation string,
	rules [][]string,
	errPtr *error,
) func() {
	if !e.shouldLog(eventType) {
		return func() {}
	}

	entry := log.NewLogEntry(eventType)
	entry.Operation = operation
	entry.Rules = rules
	entry.RuleCount = len(rules)

	handle := e.logger.OnBeforeEvent(entry)

	return func() {
		entry.Duration = time.Since(entry.Timestamp)
		if errPtr != nil {
			entry.Error = *errPtr
		}
		e.logger.OnAfterEvent(handle, entry)
	}
}

// LogPolicyEventWithCount records a policy-related event with a count setter.
// Usage:
//
//	var err error
//	var count int
//	defer e.LogPolicyEventWithCount(log.EventPolicyLoad, "load", nil, &err, &count)()
//	// ... do policy operation ...
//	count = actualCount
func (e *Enforcer) LogPolicyEventWithCount(
	eventType log.EventType,
	operation string,
	rules [][]string,
	errPtr *error,
	countPtr *int,
) func() {
	if !e.shouldLog(eventType) {
		return func() {}
	}

	entry := log.NewLogEntry(eventType)
	entry.Operation = operation
	entry.Rules = rules

	handle := e.logger.OnBeforeEvent(entry)

	return func() {
		entry.Duration = time.Since(entry.Timestamp)
		if errPtr != nil {
			entry.Error = *errPtr
		}
		if countPtr != nil {
			entry.RuleCount = *countPtr
		} else if rules != nil {
			entry.RuleCount = len(rules)
		}
		e.logger.OnAfterEvent(handle, entry)
	}
}

// LogRoleEvent records a role-related event.
// Usage:
//
//	var err error
//	defer e.LogRoleEvent(log.EventRoleAdd, [][]string{{"alice", "admin"}}, &err)()
//	// ... do role operation ...
//	err = actualError
func (e *Enforcer) LogRoleEvent(
	eventType log.EventType,
	rules [][]string,
	errPtr *error,
) func() {
	if !e.shouldLog(eventType) {
		return func() {}
	}

	entry := log.NewLogEntry(eventType)
	entry.Operation = string(eventType)
	entry.Rules = rules
	entry.RuleCount = len(rules)

	handle := e.logger.OnBeforeEvent(entry)

	return func() {
		entry.Duration = time.Since(entry.Timestamp)
		if errPtr != nil {
			entry.Error = *errPtr
		}
		e.logger.OnAfterEvent(handle, entry)
	}
}

// LogModelEvent records a model-related event.
// Usage:
//
//	var err error
//	defer e.LogModelEvent(log.EventModelLoad, &err)()
//	// ... do model operation ...
//	err = actualError
func (e *Enforcer) LogModelEvent(eventType log.EventType, errPtr *error) func() {
	if !e.shouldLog(eventType) {
		return func() {}
	}

	entry := log.NewLogEntry(eventType)
	entry.Operation = string(eventType)

	handle := e.logger.OnBeforeEvent(entry)

	return func() {
		entry.Duration = time.Since(entry.Timestamp)
		if errPtr != nil {
			entry.Error = *errPtr
		}
		e.logger.OnAfterEvent(handle, entry)
	}
}
