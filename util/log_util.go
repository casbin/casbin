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

package util

import "log"

// EnableLog controls whether to print log to console.
var EnableLog = true

// Logger is the logging interface implementation.
type Logger interface {
	//Print formats using the default formats for its operands and logs the message.
	Print(...interface{})

	//Printf formats according to a format specifier and logs the message.
	Printf(string, ...interface{})
}

// DefaultLogger is the implementation for a Logger using golang log.
type DefaultLogger struct{}

func (l *DefaultLogger) Print(v ...interface{})                 { log.Print(v...) }
func (l *DefaultLogger) Printf(format string, v ...interface{}) { log.Printf(format, v...) }

var logger Logger = &DefaultLogger{}

// SetLogger sets the current logger.
func SetLogger(l Logger) {
	logger = l
}

// LogPrint prints the log.
func LogPrint(v ...interface{}) {
	if EnableLog {
		logger.Print(v...)
	}
}

// LogPrintf prints the log with the format.
func LogPrintf(format string, v ...interface{}) {
	if EnableLog {
		logger.Printf(format, v...)
	}
}
