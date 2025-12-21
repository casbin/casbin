// Copyright 2024 The casbin Authors. All Rights Reserved.
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

// DefaultLogger is a no-op logger implementation.
type DefaultLogger struct {
	enabled   bool
	subscribe []EventType
}

// NewDefaultLogger creates a new DefaultLogger.
func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{
		enabled:   false,
		subscribe: nil, // nil means subscribe to all events.
	}
}

// Enable turns the logger on or off.
func (l *DefaultLogger) Enable(enabled bool) {
	l.enabled = enabled
}

// IsEnabled returns whether the logger is enabled.
func (l *DefaultLogger) IsEnabled() bool {
	return l.enabled
}

// Subscribe returns the list of event types this logger is interested in.
func (l *DefaultLogger) Subscribe() []EventType {
	return l.subscribe
}

// OnBeforeEvent is called before an event occurs.
func (l *DefaultLogger) OnBeforeEvent(entry *LogEntry) *Handle {
	return NewHandle()
}

// OnAfterEvent is called after an event completes.
func (l *DefaultLogger) OnAfterEvent(handle *Handle, entry *LogEntry) {
	// Default implementation does nothing.
}
