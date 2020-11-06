// Copyright 2020 The casbin Authors. All Rights Reserved.
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

package casbin

import (
	"context"
	"errors"

	"github.com/casbin/casbin/persist"
)

// ContextEnforcer wraps Enforcer and provides context handling
type ContextEnforcer struct {
	*Enforcer
}

// NewCachedEnforcer creates a cached enforcer via file or DB.
func NewContextEnforcer(params ...interface{}) (*ContextEnforcer, error) {
	e := &ContextEnforcer{}

	var err error

	e.Enforcer, err = NewEnforcer(params...)
	if err != nil {
		return nil, err
	}

	return e, nil
}

func (e *ContextEnforcer) LoadPolicy(ctx context.Context) error {
	var ctxAdapter persist.ContextAdapter

	// Attempt to cast the Adapter as a ContextAdapter
	switch adapter := e.adapter.(type) {
	case persist.ContextAdapter:
		ctxAdapter = adapter
	default:
		return errors.New("context methods are not supported by this adapter")
	}

	e.model.ClearPolicy()

	if err := ctxAdapter.LoadPolicyWIthContext(ctx, e.model); err != nil && err.Error() != "invalid file path, file path cannot be empty" {
		return err
	}

	if e.autoBuildRoleLinks {
		return e.BuildRoleLinks()
	}

	return nil
}

func (e *ContextEnforcer) loadFilteredPolicy(ctx context.Context, filter interface{}) error {
	var ctxAdapter persist.FilteredContextAdapter

	// Attempt to cast the Adapter as a FilteredContextAdapter
	switch adapter := e.adapter.(type) {
	case persist.FilteredContextAdapter:
		ctxAdapter = adapter
	default:
		return errors.New("filtered context methods are not supported by this adapter")
	}

	if err := ctxAdapter.LoadFilteredPolicyWithContext(ctx, e.model, filter); err != nil && err.Error() != "invalid file path, file path cannot be empty" {
		return err
	}

	e.model.PrintPolicy()

	if e.autoBuildRoleLinks {
		return e.BuildRoleLinks()
	}

	return nil
}

func (e *ContextEnforcer) LoadFilteredPolicy(ctx context.Context, filter interface{}) error {
	e.model.ClearPolicy()

	return e.loadFilteredPolicy(ctx, filter)
}

func (e *ContextEnforcer) LoadIncrementalFilteredPolicy(ctx context.Context, filter interface{}) error {
	return e.loadFilteredPolicy(ctx, filter)
}
