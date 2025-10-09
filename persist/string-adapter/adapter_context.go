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

package stringadapter

import (
	"context"

	"github.com/casbin/casbin/v2/model"
)

// LoadPolicyCtx loads all policy rules from the storage with context.
func (a *Adapter) LoadPolicyCtx(ctx context.Context, model model.Model) error {
	if err := checkCtx(ctx); err != nil {
		return err
	}

	return a.LoadPolicy(model)
}

// SavePolicyCtx saves all policy rules to the storage with context.
func (a *Adapter) SavePolicyCtx(ctx context.Context, model model.Model) error {
	if err := checkCtx(ctx); err != nil {
		return err
	}

	return a.SavePolicy(model)
}

// AddPolicyCtx adds a policy rule to the storage with context.
func (a *Adapter) AddPolicyCtx(ctx context.Context, sec string, ptype string, rule []string) error {
	if err := checkCtx(ctx); err != nil {
		return err
	}

	return a.AddPolicy(sec, ptype, rule)
}

// RemovePolicyCtx removes a policy rule from the storage with context.
func (a *Adapter) RemovePolicyCtx(ctx context.Context, sec string, ptype string, rule []string) error {
	if err := checkCtx(ctx); err != nil {
		return err
	}

	return a.RemovePolicy(sec, ptype, rule)
}

// RemoveFilteredPolicyCtx removes policy rules that match the filter from the storage with context.
func (a *Adapter) RemoveFilteredPolicyCtx(ctx context.Context, sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	if err := checkCtx(ctx); err != nil {
		return err
	}

	return a.RemoveFilteredPolicy(sec, ptype, fieldIndex, fieldValues...)
}

func checkCtx(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
