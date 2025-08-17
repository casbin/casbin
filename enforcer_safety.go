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

package casbin

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// SafeEnforcer wraps Enforcer and provides additional safety features.
type SafeEnforcer struct {
	*Enforcer
	defaultTimeout time.Duration
}

// NewSafeEnforcer creates a safe enforcer with default timeout of 5 seconds.
func NewSafeEnforcer(params ...interface{}) (*SafeEnforcer, error) {
	e, err := NewEnforcer(params...)
	if err != nil {
		return nil, err
	}
	
	return &SafeEnforcer{
		Enforcer:       e,
		defaultTimeout: 5 * time.Second,
	}, nil
}

// SetDefaultTimeout sets the default timeout for operations.
func (e *SafeEnforcer) SetDefaultTimeout(timeout time.Duration) {
	e.defaultTimeout = timeout
}

// EnforceWithTimeout decides whether a "subject" can access a "object" with the operation "action" with a timeout.
func (e *SafeEnforcer) EnforceWithTimeout(timeout time.Duration, rvals ...interface{}) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	resultChan := make(chan bool, 1)
	errChan := make(chan error, 1)
	
	go func() {
		result, err := e.Enforcer.Enforce(rvals...)
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- result
	}()
	
	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errChan:
		return false, err
	case <-ctx.Done():
		return false, errors.New("enforce operation timed out")
	}
}

// SafeEnforce decides whether a "subject" can access a "object" with the operation "action" with the default timeout.
func (e *SafeEnforcer) SafeEnforce(rvals ...interface{}) (bool, error) {
	return e.EnforceWithTimeout(e.defaultTimeout, rvals...)
}

// LoadPolicyWithTimeout reloads the policy from file/database with a timeout.
func (e *SafeEnforcer) LoadPolicyWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	errChan := make(chan error, 1)
	
	go func() {
		errChan <- e.Enforcer.LoadPolicy()
	}()
	
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return errors.New("load policy operation timed out")
	}
}

// SafeLoadPolicy reloads the policy from file/database with the default timeout.
func (e *SafeEnforcer) SafeLoadPolicy() error {
	return e.LoadPolicyWithTimeout(e.defaultTimeout)
}

// ValidateEnforceParams validates the parameters for Enforce.
func (e *SafeEnforcer) ValidateEnforceParams(rvals ...interface{}) error {
	if len(rvals) == 0 {
		return errors.New("missing parameters for enforcement")
	}
	
	if len(e.model["r"]) == 0 || len(e.model["r"]["r"].Tokens) == 0 {
		return errors.New("model not initialized properly")
	}
	
	if len(rvals) != len(e.model["r"]["r"].Tokens) {
		return fmt.Errorf("invalid request size: expected %d, got %d", 
			len(e.model["r"]["r"].Tokens), len(rvals))
	}
	
	return nil
}

// EnforceWithValidation validates parameters before enforcement.
func (e *SafeEnforcer) EnforceWithValidation(rvals ...interface{}) (bool, error) {
	if err := e.ValidateEnforceParams(rvals...); err != nil {
		return false, err
	}
	
	return e.Enforcer.Enforce(rvals...)
}

// SafeEnforceWithMatcher validates parameters before enforcement with a custom matcher.
func (e *SafeEnforcer) SafeEnforceWithMatcher(matcher string, rvals ...interface{}) (bool, error) {
	if err := e.ValidateEnforceParams(rvals...); err != nil {
		return false, err
	}
	
	return e.Enforcer.EnforceWithMatcher(matcher, rvals...)
}

// RecoverableEnforce wraps Enforce with panic recovery.
func (e *SafeEnforcer) RecoverableEnforce(rvals ...interface{}) (result bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			result = false
			err = fmt.Errorf("panic in enforce operation: %v", r)
		}
	}()
	
	return e.Enforcer.Enforce(rvals...)
}

// HealthCheck performs a basic health check on the enforcer.
func (e *SafeEnforcer) HealthCheck() error {
	if e.model == nil {
		return errors.New("model is nil")
	}
	
	if e.rmMap == nil {
		return errors.New("role manager map is nil")
	}
	
	if e.eft == nil {
		return errors.New("effector is nil")
	}
	
	return nil
}