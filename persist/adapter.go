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

package persist

import (
	"encoding/csv"
	"strings"

	"github.com/casbin/casbin/v2/model"
)

// LoadPolicyLine loads a text line as a policy rule to model.
func LoadPolicyLine(line string, m model.Model) error {
	if line == "" || strings.HasPrefix(line, "#") {
		return nil
	}

	r := csv.NewReader(strings.NewReader(line))
	r.Comma = ','
	r.Comment = '#'
	r.TrimLeadingSpace = true

	tokens, err := r.Read()
	if err != nil {
		return err
	}

	return LoadPolicyArray(tokens, m)
}

// LoadPolicyArray loads a policy rule to model.
func LoadPolicyArray(rule []string, m model.Model) error {
	key := rule[0]
	sec := key[:1]
	ok, err := m.HasPolicyEx(sec, key, rule[1:])
	if err != nil {
		return err
	}
	if ok {
		return nil // skip duplicated policy
	}

	err = m.AddPolicy(sec, key, rule[1:])
	if err != nil {
		return err
	}

	return nil
}

// NullableString represents a string value that can be stored in databases.
// It is designed to preserve empty strings as distinct from NULL values.
// This is important because empty strings have semantic meaning in policies
// (e.g., to denote omitted/irrelevant fields), which is different from NULL
// or wildcard values.
//
// Usage for database adapters:
//   - When saving to database: use StringToNullable(value)
//   - When loading from database: use NullableToString(dbValue, valid)
type NullableString struct {
	Value string
	Valid bool
}

// StringToNullable converts a string to NullableString for database storage.
// Empty strings are preserved with Valid=true, ensuring they are stored as
// empty strings rather than NULL in the database.
// This function should be used by database adapters when storing policy rules.
func StringToNullable(s string) NullableString {
	// Empty strings are valid and should be preserved as empty strings, not NULL
	return NullableString{
		Value: s,
		Valid: true,
	}
}

// NullableToString converts a database value to string for policy loading.
// If valid is false (NULL in database), it returns an empty string.
// If valid is true, it returns the actual value (which may be an empty string).
// This function should be used by database adapters when loading policy rules.
func NullableToString(value string, valid bool) string {
	if !valid {
		// NULL from database becomes empty string
		return ""
	}
	return value
}

// Adapter is the interface for Casbin adapters.
type Adapter interface {
	// LoadPolicy loads all policy rules from the storage.
	LoadPolicy(model model.Model) error
	// SavePolicy saves all policy rules to the storage.
	SavePolicy(model model.Model) error

	// AddPolicy adds a policy rule to the storage.
	// This is part of the Auto-Save feature.
	AddPolicy(sec string, ptype string, rule []string) error
	// RemovePolicy removes a policy rule from the storage.
	// This is part of the Auto-Save feature.
	RemovePolicy(sec string, ptype string, rule []string) error
	// RemoveFilteredPolicy removes policy rules that match the filter from the storage.
	// This is part of the Auto-Save feature.
	RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error
}
