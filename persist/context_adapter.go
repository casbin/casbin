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
	"context"

	"github.com/casbin/casbin/v2/model"
)

type ContextAdapter interface {
	Adapter

	// LoadPolicyWithContext loads all policy rules from the storage.
	LoadPolicyWithContext(ctx context.Context, model model.Model) error
	// SavePolicyWithContext saves all policy rules to the storage.
	SavePolicyWithContext(ctx context.Context, model model.Model) error

	// AddPolicyWithContext adds a policy rule to the storage.
	// This is part of the Auto-Save feature.
	AddPolicyWithContext(ctx context.Context, sec string, ptype string, rule []string) error
	// RemovePolicyWithContext removes a policy rule from the storage.
	// This is part of the Auto-Save feature.
	RemovePolicyWithContext(ctx context.Context, sec string, ptype string, rule []string) error
	// RemoveFilteredPolicyWithContext removes policy rules that match the filter from the storage.
	// This is part of the Auto-Save feature.
	RemoveFilteredPolicyWithContext(ctx context.Context, sec string, ptype string, fieldIndex int, fieldValues ...string) error
}
