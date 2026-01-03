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

package log

import (
	"fmt"
	"io"
	"os"
	"time"
)

// DefaultLogger is the default implementation of the Logger interface.
type DefaultLogger struct {
	output      io.Writer
	eventTypes  map[EventType]bool
	logCallback func(entry *LogEntry) error
}

// NewDefaultLogger creates a new DefaultLogger instance.
// If no output is set via SetOutput, it defaults to os.Stdout.
func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{
		output:     os.Stdout,
		eventTypes: make(map[EventType]bool),
	}
}

// SetOutput sets the output destination for the logger.
// It can be set to a buffer or any io.Writer.
func (l *DefaultLogger) SetOutput(w io.Writer) {
	if w != nil {
		l.output = w
	}
}

// SetEventTypes sets the event types that should be logged.
// Only events matching these types will have IsActive set to true.
func (l *DefaultLogger) SetEventTypes(eventTypes []EventType) error {
	l.eventTypes = make(map[EventType]bool)
	for _, et := range eventTypes {
		l.eventTypes[et] = true
	}
	return nil
}

// OnBeforeEvent is called before an event occurs.
// It sets the StartTime and determines if the event should be active based on configured event types.
func (l *DefaultLogger) OnBeforeEvent(entry *LogEntry) error {
	if entry == nil {
		return fmt.Errorf("log entry is nil")
	}

	entry.StartTime = time.Now()

	// Set IsActive based on whether this event type is enabled
	// If no event types are configured, all events are considered active
	if len(l.eventTypes) == 0 {
		entry.IsActive = true
	} else {
		entry.IsActive = l.eventTypes[entry.EventType]
	}

	return nil
}

// OnAfterEvent is called after an event completes.
// It calculates the duration, logs the entry if active, and calls the user callback if set.
func (l *DefaultLogger) OnAfterEvent(entry *LogEntry) error {
	if entry == nil {
		return fmt.Errorf("log entry is nil")
	}

	entry.EndTime = time.Now()
	entry.Duration = entry.EndTime.Sub(entry.StartTime)

	// Only log if the event is active
	if entry.IsActive && l.output != nil {
		if err := l.writeLog(entry); err != nil {
			return err
		}
	}

	// Call user-provided callback if set
	if l.logCallback != nil {
		if err := l.logCallback(entry); err != nil {
			return err
		}
	}

	return nil
}

// SetLogCallback sets a user-provided callback function.
// The callback is called at the end of OnAfterEvent.
func (l *DefaultLogger) SetLogCallback(callback func(entry *LogEntry) error) error {
	l.logCallback = callback
	return nil
}

// writeLog writes the log entry to the configured output.
func (l *DefaultLogger) writeLog(entry *LogEntry) error {
	var logMessage string

	switch entry.EventType {
	case EventEnforce:
		logMessage = fmt.Sprintf("[%s] Enforce: subject=%s, object=%s, action=%s, domain=%s, allowed=%v, duration=%v\n",
			entry.EventType, entry.Subject, entry.Object, entry.Action, entry.Domain, entry.Allowed, entry.Duration)
	case EventAddPolicy, EventRemovePolicy:
		logMessage = fmt.Sprintf("[%s] RuleCount=%d, duration=%v\n",
			entry.EventType, entry.RuleCount, entry.Duration)
	case EventLoadPolicy, EventSavePolicy:
		logMessage = fmt.Sprintf("[%s] RuleCount=%d, duration=%v\n",
			entry.EventType, entry.RuleCount, entry.Duration)
	default:
		logMessage = fmt.Sprintf("[%s] duration=%v\n",
			entry.EventType, entry.Duration)
	}

	if entry.Error != nil {
		logMessage = fmt.Sprintf("%s Error: %v\n", logMessage[:len(logMessage)-1], entry.Error)
	}

	_, err := l.output.Write([]byte(logMessage))
	return err
}
