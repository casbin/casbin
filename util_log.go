// Copyright 2026 The casbin Authors. All Rights Reserved.
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
	"github.com/casbin/casbin/v3/log"
	"github.com/casbin/casbin/v3/model"
)

// onLogBeforeEvent calls OnBeforeEvent on the logger if it exists.
func (e *Enforcer) onLogBeforeEvent(eventType log.EventType) *log.LogEntry {
	if e.logger == nil {
		return nil
	}
	logEntry := &log.LogEntry{
		EventType: eventType,
	}
	_ = e.logger.OnBeforeEvent(logEntry)
	return logEntry
}

// onLogAfterEvent calls OnAfterEvent on the logger if it exists and logEntry is not nil.
func (e *Enforcer) onLogAfterEvent(logEntry *log.LogEntry) {
	if e.logger != nil && logEntry != nil {
		_ = e.logger.OnAfterEvent(logEntry)
	}
}

// onLogAfterEventWithError calls OnAfterEvent with an error if logger and logEntry exist.
func (e *Enforcer) onLogAfterEventWithError(logEntry *log.LogEntry, err error) {
	if e.logger != nil && logEntry != nil {
		logEntry.Error = err
		_ = e.logger.OnAfterEvent(logEntry)
	}
}

// countModelRules counts the total number of rules in a model's p and g sections.
func countModelRules(m model.Model) int {
	ruleCount := 0
	if pSection, ok := m["p"]; ok {
		for _, ast := range pSection {
			ruleCount += len(ast.Policy)
		}
	}
	if gSection, ok := m["g"]; ok {
		for _, ast := range gSection {
			ruleCount += len(ast.Policy)
		}
	}
	return ruleCount
}

// onLogBeforeEventInLoadPolicy initializes logging for LoadPolicy operation.
func (e *Enforcer) onLogBeforeEventInLoadPolicy() *log.LogEntry {
	return e.onLogBeforeEvent(log.EventLoadPolicy)
}

// onLogAfterEventInLoadPolicy finalizes logging for LoadPolicy operation with rule count.
func (e *Enforcer) onLogAfterEventInLoadPolicy(logEntry *log.LogEntry, newModel model.Model) {
	if e.logger != nil && logEntry != nil {
		logEntry.RuleCount = countModelRules(newModel)
		_ = e.logger.OnAfterEvent(logEntry)
	}
}

// onLogBeforeEventInSavePolicy initializes logging for SavePolicy operation with rule count.
func (e *Enforcer) onLogBeforeEventInSavePolicy() *log.LogEntry {
	if e.logger == nil {
		return nil
	}
	logEntry := &log.LogEntry{
		EventType: log.EventSavePolicy,
		RuleCount: countModelRules(e.model),
	}
	_ = e.logger.OnBeforeEvent(logEntry)
	return logEntry
}

// onLogAfterEventInSavePolicy finalizes logging for SavePolicy operation.
func (e *Enforcer) onLogAfterEventInSavePolicy(logEntry *log.LogEntry) {
	e.onLogAfterEvent(logEntry)
}

// createEnforceLogEntry creates a log entry for enforce events with subject, object, action, and domain extracted from rvals.
func (e *Enforcer) createEnforceLogEntry(rvals []interface{}) *log.LogEntry {
	entry := &log.LogEntry{
		EventType: log.EventEnforce,
	}
	if len(rvals) > 0 {
		if s, isString := rvals[0].(string); isString {
			entry.Subject = s
		}
	}
	if len(rvals) > 1 {
		if o, isString := rvals[1].(string); isString {
			entry.Object = o
		}
	}
	if len(rvals) > 2 {
		if a, isString := rvals[2].(string); isString {
			entry.Action = a
		}
	}
	if len(rvals) > 3 {
		if d, isString := rvals[3].(string); isString {
			entry.Domain = d
		}
	}
	return entry
}

// onLogBeforeEventInEnforce initializes logging for Enforce operation.
func (e *Enforcer) onLogBeforeEventInEnforce(rvals []interface{}) *log.LogEntry {
	if e.logger == nil {
		return nil
	}
	logEntry := e.createEnforceLogEntry(rvals)
	_ = e.logger.OnBeforeEvent(logEntry)
	return logEntry
}

// onLogAfterEventInEnforce finalizes logging for Enforce operation.
func (e *Enforcer) onLogAfterEventInEnforce(logEntry *log.LogEntry, allowed bool) {
	if e.logger != nil && logEntry != nil {
		logEntry.Allowed = allowed
		_ = e.logger.OnAfterEvent(logEntry)
	}
}

// logPolicyOperation logs a policy operation (add or remove) with before and after events.
func (e *Enforcer) logPolicyOperation(eventType log.EventType, sec string, rule []string, operation func() (bool, error)) (bool, error) {
	var logEntry *log.LogEntry
	if e.logger != nil && sec == "p" {
		logEntry = &log.LogEntry{
			EventType: eventType,
			Rules:     [][]string{rule},
		}
		_ = e.logger.OnBeforeEvent(logEntry)
	}

	ok, err := operation()

	if e.logger != nil && logEntry != nil {
		if ok && err == nil {
			logEntry.RuleCount = 1
		} else {
			logEntry.RuleCount = 0
			logEntry.Error = err
		}
		_ = e.logger.OnAfterEvent(logEntry)
	}

	return ok, err
}
