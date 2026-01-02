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

import (
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"
)

// DefaultLogger is the default implementation of the Logger interface.
// It outputs structured JSON logs to the configured writer.
type DefaultLogger struct {
	mu            sync.RWMutex
	enabled       bool
	subscriptions []EventType
	writer        io.Writer
}

// NewDefaultLogger creates a new DefaultLogger that writes to the given writer.
// If writer is nil, it defaults to os.Stdout.
// If subscriptions is nil or empty, it subscribes to all events.
func NewDefaultLogger(writer io.Writer, subscriptions []EventType) *DefaultLogger {
	if writer == nil {
		writer = os.Stdout
	}
	return &DefaultLogger{
		enabled:       true,
		subscriptions: subscriptions,
		writer:        writer,
	}
}

// Enable enables or disables the logger in a thread-safe manner.
func (l *DefaultLogger) Enable(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enabled = enabled
}

// IsEnabled returns whether the logger is currently enabled.
func (l *DefaultLogger) IsEnabled() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.enabled
}

// Subscribe returns the list of event types this logger is subscribed to.
// Returning nil means all events are subscribed.
func (l *DefaultLogger) Subscribe() []EventType {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.subscriptions
}

// OnBeforeEvent is called before an event occurs and returns a handle for context.
// It initializes a new Handle with the current time.
func (l *DefaultLogger) OnBeforeEvent(entry *LogEntry) *Handle {
	if !l.IsEnabled() {
		return nil
	}
	return &Handle{
		StartTime: time.Now(),
		Store:     make(map[string]interface{}),
	}
}

// OnAfterEvent is called after an event completes with the handle and final entry.
// It calculates the duration, formats the LogEntry into a map, and outputs it as JSON.
func (l *DefaultLogger) OnAfterEvent(handle *Handle, entry *LogEntry) {
	if !l.IsEnabled() || handle == nil || entry == nil {
		return
	}

	// Calculate duration
	entry.Duration = time.Since(handle.StartTime)

	// Convert LogEntry to map for JSON serialization
	logMap := l.entryToMap(entry)

	// Serialize to JSON
	data, err := json.Marshal(logMap)
	if err != nil {
		return
	}

	// Write with lock protection to ensure thread-safety
	l.mu.Lock()
	defer l.mu.Unlock()
	l.writer.Write(data)
	l.writer.Write([]byte("\n"))
}

// entryToMap converts a LogEntry to a map for JSON serialization.
func (l *DefaultLogger) entryToMap(entry *LogEntry) map[string]interface{} {
	result := make(map[string]interface{})

	// Always include basic fields
	result["event_type"] = string(entry.EventType)
	result["timestamp"] = entry.Timestamp.Format(time.RFC3339Nano)
	result["duration_ms"] = float64(entry.Duration.Microseconds()) / 1000.0

	// Include enforce-specific fields
	if entry.Subject != "" {
		result["subject"] = entry.Subject
	}
	if entry.Object != "" {
		result["object"] = entry.Object
	}
	if entry.Action != "" {
		result["action"] = entry.Action
	}
	if entry.Domain != "" {
		result["domain"] = entry.Domain
	}
	if entry.EventType == EventEnforce {
		result["allowed"] = entry.Allowed
	}
	if entry.Matched != "" {
		result["matched"] = entry.Matched
	}

	// Include policy operation fields
	if entry.Operation != "" {
		result["operation"] = entry.Operation
	}
	if entry.Rules != nil && len(entry.Rules) > 0 {
		result["rules"] = entry.Rules
	}
	if entry.RuleCount > 0 {
		result["rule_count"] = entry.RuleCount
	}

	// Include error if present
	if entry.Error != nil {
		result["error"] = entry.Error.Error()
	}

	// Include additional attributes
	if entry.Attributes != nil && len(entry.Attributes) > 0 {
		for k, v := range entry.Attributes {
			result[k] = v
		}
	}

	return result
}
