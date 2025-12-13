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

// DefaultEventLogger is a no-op event logger implementation
type DefaultEventLogger struct {
	enabled   bool
	subscribe []EventType
}

// NewDefaultEventLogger creates a new DefaultEventLogger
func NewDefaultEventLogger() *DefaultEventLogger {
	return &DefaultEventLogger{
		enabled:   false,
		subscribe: nil, // nil means subscribe to all events
	}
}

func (l *DefaultEventLogger) Enable(enabled bool) {
	l.enabled = enabled
}

func (l *DefaultEventLogger) IsEnabled() bool {
	return l.enabled
}

func (l *DefaultEventLogger) Subscribe() []EventType {
	return l.subscribe
}

func (l *DefaultEventLogger) OnBeforeEvent(entry *LogEntry) *Handle {
	return NewHandle()
}

func (l *DefaultEventLogger) OnAfterEvent(handle *Handle, entry *LogEntry) {
	// Default implementation does nothing
}
