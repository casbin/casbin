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
	SetEventTypes([]EventType) error
	// OnBeforeEvent is called before an event occurs and returns a handle for context.
	OnBeforeEvent(entry *LogEntry) error
	// OnAfterEvent is called after an event completes with the handle and final entry.
	OnAfterEvent(entry *LogEntry) error

	SetLogCallback(func(entry *LogEntry) error) error
}
