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

package fileadapter

import (
	"context"

	"github.com/casbin/casbin/v2/model"
)

// LoadPolicyCtx loads all policy rules from the storage with context.
func (a *FilteredAdapter) LoadPolicyCtx(ctx context.Context, model model.Model) error {
	if err := checkCtx(ctx); err != nil {
		return err
	}

	return a.LoadPolicy(model)
}

// LoadFilteredPolicyCtx loads only policy rules that match the filter with context.
func (a *FilteredAdapter) LoadFilteredPolicyCtx(ctx context.Context, model model.Model, filter interface{}) error {
	if err := checkCtx(ctx); err != nil {
		return err
	}

	return a.LoadFilteredPolicy(model, filter)
}

// SavePolicyCtx saves all policy rules to the storage with context.
func (a *FilteredAdapter) SavePolicyCtx(ctx context.Context, model model.Model) error {
	if err := checkCtx(ctx); err != nil {
		return err
	}

	return a.SavePolicy(model)
}

// IsFilteredCtx returns true if the loaded policy has been filtered with context.
func (a *FilteredAdapter) IsFilteredCtx(ctx context.Context) bool {
	if err := checkCtx(ctx); err != nil {
		return false
	}

	return a.IsFiltered()
}
