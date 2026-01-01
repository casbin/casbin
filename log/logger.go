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

// Logger defines the interface for event-driven logging in Casbin.
type Logger interface {
	// Enable enables or disables the logger.
	Enable(enabled bool)
	// IsEnabled returns whether the logger is currently enabled.
	IsEnabled() bool
	// Subscribe returns the list of event types this logger is subscribed to.
	Subscribe() []EventType
	// OnBeforeEvent is called before an event occurs and returns a handle for context.
	OnBeforeEvent(entry *LogEntry) *Handle
	// OnAfterEvent is called after an event completes with the handle and final entry.
	OnAfterEvent(handle *Handle, entry *LogEntry)
}
