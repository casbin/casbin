// Copyright 2023 The casbin Authors. All Rights Reserved.
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
	"context"

	"github.com/casbin/casbin/v2/model"
)

// ContextAdapter provides a context-aware interface for Casbin adapters.
type ContextAdapter interface {
	Adapter

	// LoadPolicy loads all policy rules from the storage with context.
	LoadPolicy(ctx context.Context, model model.Model) error
	// SavePolicy saves all policy rules to the storage with context.
	SavePolicy(ctx context.Context, model model.Model) error

	// AddPolicy adds a policy rule to the storage with context.
	// This is part of the Auto-Save feature.
	AddPolicy(ctx context.Context, sec string, ptype string, rule []string) error
	// RemovePolicy removes a policy rule from the storage with context.
	// This is part of the Auto-Save feature.
	RemovePolicy(ctx context.Context, sec string, ptype string, rule []string) error
	// RemoveFilteredPolicy removes policy rules that match the filter from the storage with context.
	// This is part of the Auto-Save feature.
	RemoveFilteredPolicy(ctx context.Context, sec string, ptype string, fieldIndex int, fieldValues ...string) error
}
