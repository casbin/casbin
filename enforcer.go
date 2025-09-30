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

package casbin

import (
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/casbin/casbin/v2/effector"
	"github.com/casbin/casbin/v2/log"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	"github.com/casbin/casbin/v2/rbac"
	defaultrolemanager "github.com/casbin/casbin/v2/rbac/default-role-manager"
	"github.com/casbin/casbin/v2/util"

	"github.com/casbin/govaluate"
)

// Enforcer is the main interface for authorization enforcement and policy management.
type Enforcer struct {
	modelPath string
	model     model.Model
	fm        model.FunctionMap
	eft       effector.Effector

	adapter    persist.Adapter
	watcher    persist.Watcher
	dispatcher persist.Dispatcher
	rmMap      map[string]rbac.RoleManager
	condRmMap  map[string]rbac.ConditionalRoleManager
	matcherMap sync.Map

	enabled              bool
	autoSave             bool
	autoBuildRoleLinks   bool
	autoNotifyWatcher    bool
	autoNotifyDispatcher bool
	acceptJsonRequest    bool

	logger log.Logger
}

// EnforceContext is used as the first element of the parameter "rvals" in method "enforce".
type EnforceContext struct {
	RType string
	PType string
	EType string
	MType string
}

func (e EnforceContext) GetCacheKey() string {
	return "EnforceContext{" + e.RType + "-" + e.PType + "-" + e.EType + "-" + e.MType + "}"
}

// EnforcementContext holds the execution context for policy evaluation.
type EnforcementContext struct {
	RType      string
	PType      string
	EType      string
	MType      string
	RTokens    map[string]int
	PTokens    map[string]int
	RVals      []interface{}
	Parameters enforceParameters
	ExpString  string
	HasEval    bool
}

// PolicyEvaluationResult contains the result of policy evaluation.
type PolicyEvaluationResult struct {
	Effect        effector.Effect
	ExplainIndex  int
	PolicyEffects []effector.Effect
	MatchResults  []float64
}

// EnforcementError represents specific error types during enforcement.
type EnforcementError struct {
	Type    EnforcementErrorType
	Message string
	Context map[string]interface{}
}

func (e EnforcementError) Error() string {
	return e.Message
}

// EnforcementErrorType represents different types of enforcement errors.
type EnforcementErrorType int

const (
	ErrInvalidRequest EnforcementErrorType = iota
	ErrInvalidPolicy
	ErrExpressionCompilation
	ErrEvaluationFailure
	ErrConfigurationError
)

// NewEnforcer creates an enforcer via file or DB.
//
// File:
//
//	e := casbin.NewEnforcer("path/to/basic_model.conf", "path/to/basic_policy.csv")
//
// MySQL DB:
//
//	a := mysqladapter.NewDBAdapter("mysql", "mysql_username:mysql_password@tcp(127.0.0.1:3306)/")
//	e := casbin.NewEnforcer("path/to/basic_model.conf", a)
func NewEnforcer(params ...interface{}) (*Enforcer, error) {
	e := &Enforcer{logger: &log.DefaultLogger{}}

	parsedParamLen := 0
	paramLen := len(params)
	if paramLen >= 1 {
		enableLog, ok := params[paramLen-1].(bool)
		if ok {
			e.EnableLog(enableLog)
			parsedParamLen++
		}
	}

	if paramLen-parsedParamLen >= 1 {
		logger, ok := params[paramLen-parsedParamLen-1].(log.Logger)
		if ok {
			e.logger = logger
			parsedParamLen++
		}
	}

	switch paramLen - parsedParamLen {
	case 2:
		switch p0 := params[0].(type) {
		case string:
			switch p1 := params[1].(type) {
			case string:
				err := e.InitWithFile(p0, p1)
				if err != nil {
					return nil, err
				}
			default:
				err := e.InitWithAdapter(p0, p1.(persist.Adapter))
				if err != nil {
					return nil, err
				}
			}
		default:
			switch params[1].(type) {
			case string:
				return nil, errors.New("invalid parameters for enforcer")
			default:
				err := e.InitWithModelAndAdapter(p0.(model.Model), params[1].(persist.Adapter))
				if err != nil {
					return nil, err
				}
			}
		}
	case 1:
		switch p0 := params[0].(type) {
		case string:
			err := e.InitWithFile(p0, "")
			if err != nil {
				return nil, err
			}
		default:
			err := e.InitWithModelAndAdapter(p0.(model.Model), nil)
			if err != nil {
				return nil, err
			}
		}
	case 0:
		return e, nil
	default:
		return nil, errors.New("invalid parameters for enforcer")
	}

	return e, nil
}

// InitWithFile initializes an enforcer with a model file and a policy file.
func (e *Enforcer) InitWithFile(modelPath string, policyPath string) error {
	a := fileadapter.NewAdapter(policyPath)
	return e.InitWithAdapter(modelPath, a)
}

// InitWithAdapter initializes an enforcer with a database adapter.
func (e *Enforcer) InitWithAdapter(modelPath string, adapter persist.Adapter) error {
	m, err := model.NewModelFromFile(modelPath)
	if err != nil {
		return err
	}

	err = e.InitWithModelAndAdapter(m, adapter)
	if err != nil {
		return err
	}

	e.modelPath = modelPath
	return nil
}

// InitWithModelAndAdapter initializes an enforcer with a model and a database adapter.
func (e *Enforcer) InitWithModelAndAdapter(m model.Model, adapter persist.Adapter) error {
	e.adapter = adapter

	e.model = m
	m.SetLogger(e.logger)
	e.model.PrintModel()
	e.fm = model.LoadFunctionMap()

	e.initialize()

	// Do not initialize the full policy when using a filtered adapter
	fa, ok := e.adapter.(persist.FilteredAdapter)
	if e.adapter != nil && (!ok || ok && !fa.IsFiltered()) {
		err := e.LoadPolicy()
		if err != nil {
			return err
		}
	}

	return nil
}

// SetLogger changes the current enforcer's logger.
func (e *Enforcer) SetLogger(logger log.Logger) {
	e.logger = logger
	e.model.SetLogger(e.logger)
	for k := range e.rmMap {
		e.rmMap[k].SetLogger(e.logger)
	}
	for k := range e.condRmMap {
		e.condRmMap[k].SetLogger(e.logger)
	}
}

func (e *Enforcer) initialize() {
	e.rmMap = map[string]rbac.RoleManager{}
	e.condRmMap = map[string]rbac.ConditionalRoleManager{}
	e.eft = effector.NewDefaultEffector()
	e.watcher = nil
	e.matcherMap = sync.Map{}

	e.enabled = true
	e.autoSave = true
	e.autoBuildRoleLinks = true
	e.autoNotifyWatcher = true
	e.autoNotifyDispatcher = true
	e.initRmMap()
}

// LoadModel reloads the model from the model CONF file.
// Because the policy is attached to a model, so the policy is invalidated and needs to be reloaded by calling LoadPolicy().
func (e *Enforcer) LoadModel() error {
	var err error
	e.model, err = model.NewModelFromFile(e.modelPath)
	if err != nil {
		return err
	}
	e.model.SetLogger(e.logger)

	e.model.PrintModel()
	e.fm = model.LoadFunctionMap()

	e.initialize()

	return nil
}

// GetModel gets the current model.
func (e *Enforcer) GetModel() model.Model {
	return e.model
}

// SetModel sets the current model.
func (e *Enforcer) SetModel(m model.Model) {
	e.model = m
	e.fm = model.LoadFunctionMap()

	e.model.SetLogger(e.logger)
	e.initialize()
}

// GetAdapter gets the current adapter.
func (e *Enforcer) GetAdapter() persist.Adapter {
	return e.adapter
}

// SetAdapter sets the current adapter.
func (e *Enforcer) SetAdapter(adapter persist.Adapter) {
	e.adapter = adapter
}

// SetWatcher sets the current watcher.
func (e *Enforcer) SetWatcher(watcher persist.Watcher) error {
	e.watcher = watcher
	if _, ok := e.watcher.(persist.WatcherEx); ok {
		// The callback of WatcherEx has no generic implementation.
		return nil
	} else {
		// In case the Watcher wants to use a customized callback function, call `SetUpdateCallback` after `SetWatcher`.
		return watcher.SetUpdateCallback(func(string) { _ = e.LoadPolicy() })
	}
}

// GetRoleManager gets the current role manager.
func (e *Enforcer) GetRoleManager() rbac.RoleManager {
	if e.rmMap != nil && e.rmMap["g"] != nil {
		return e.rmMap["g"]
	} else if e.condRmMap != nil && e.condRmMap["g"] != nil {
		return e.condRmMap["g"]
	} else {
		return nil
	}
}

// GetNamedRoleManager gets the role manager for the named policy.
func (e *Enforcer) GetNamedRoleManager(ptype string) rbac.RoleManager {
	if e.rmMap != nil && e.rmMap[ptype] != nil {
		return e.rmMap[ptype]
	} else if e.condRmMap != nil && e.condRmMap[ptype] != nil {
		return e.condRmMap[ptype]
	} else {
		return nil
	}
}

// SetRoleManager sets the current role manager.
func (e *Enforcer) SetRoleManager(rm rbac.RoleManager) {
	e.invalidateMatcherMap()
	e.rmMap["g"] = rm
}

// SetNamedRoleManager sets the role manager for the named policy.
func (e *Enforcer) SetNamedRoleManager(ptype string, rm rbac.RoleManager) {
	e.invalidateMatcherMap()
	e.rmMap[ptype] = rm
}

// SetEffector sets the current effector.
func (e *Enforcer) SetEffector(eft effector.Effector) {
	e.eft = eft
}

// ClearPolicy clears all policy.
func (e *Enforcer) ClearPolicy() {
	e.invalidateMatcherMap()

	if e.dispatcher != nil && e.autoNotifyDispatcher {
		_ = e.dispatcher.ClearPolicy()
		return
	}
	e.model.ClearPolicy()
}

// LoadPolicy reloads the policy from file/database.
func (e *Enforcer) LoadPolicy() error {
	newModel, err := e.loadPolicyFromAdapter(e.model)
	if err != nil {
		return err
	}
	err = e.applyModifiedModel(newModel)
	if err != nil {
		return err
	}
	return nil
}

func (e *Enforcer) loadPolicyFromAdapter(baseModel model.Model) (model.Model, error) {
	newModel := baseModel.Copy()
	newModel.ClearPolicy()

	if err := e.adapter.LoadPolicy(newModel); err != nil && err.Error() != "invalid file path, file path cannot be empty" {
		return nil, err
	}

	if err := newModel.SortPoliciesBySubjectHierarchy(); err != nil {
		return nil, err
	}

	if err := newModel.SortPoliciesByPriority(); err != nil {
		return nil, err
	}

	return newModel, nil
}

func (e *Enforcer) applyModifiedModel(newModel model.Model) error {
	var err error
	needToRebuild := false
	defer func() {
		if err != nil {
			if e.autoBuildRoleLinks && needToRebuild {
				_ = e.BuildRoleLinks()
			}
		}
	}()

	if e.autoBuildRoleLinks {
		needToRebuild = true

		if err := e.rebuildRoleLinks(newModel); err != nil {
			return err
		}

		if err := e.rebuildConditionalRoleLinks(newModel); err != nil {
			return err
		}
	}

	e.model = newModel
	e.invalidateMatcherMap()
	return nil
}

func (e *Enforcer) rebuildRoleLinks(newModel model.Model) error {
	if len(e.rmMap) != 0 {
		for _, rm := range e.rmMap {
			err := rm.Clear()
			if err != nil {
				return err
			}
		}

		err := newModel.BuildRoleLinks(e.rmMap)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Enforcer) rebuildConditionalRoleLinks(newModel model.Model) error {
	if len(e.condRmMap) != 0 {
		for _, crm := range e.condRmMap {
			err := crm.Clear()
			if err != nil {
				return err
			}
		}

		err := newModel.BuildConditionalRoleLinks(e.condRmMap)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Enforcer) loadFilteredPolicy(filter interface{}) error {
	e.invalidateMatcherMap()

	var filteredAdapter persist.FilteredAdapter

	// Attempt to cast the Adapter as a FilteredAdapter
	switch adapter := e.adapter.(type) {
	case persist.FilteredAdapter:
		filteredAdapter = adapter
	default:
		return errors.New("filtered policies are not supported by this adapter")
	}
	if err := filteredAdapter.LoadFilteredPolicy(e.model, filter); err != nil && err.Error() != "invalid file path, file path cannot be empty" {
		return err
	}

	if err := e.model.SortPoliciesBySubjectHierarchy(); err != nil {
		return err
	}

	if err := e.model.SortPoliciesByPriority(); err != nil {
		return err
	}

	e.initRmMap()
	e.model.PrintPolicy()
	if e.autoBuildRoleLinks {
		err := e.BuildRoleLinks()
		if err != nil {
			return err
		}
	}
	return nil
}

// LoadFilteredPolicy reloads a filtered policy from file/database.
func (e *Enforcer) LoadFilteredPolicy(filter interface{}) error {
	e.model.ClearPolicy()

	return e.loadFilteredPolicy(filter)
}

// LoadIncrementalFilteredPolicy append a filtered policy from file/database.
func (e *Enforcer) LoadIncrementalFilteredPolicy(filter interface{}) error {
	return e.loadFilteredPolicy(filter)
}

// IsFiltered returns true if the loaded policy has been filtered.
func (e *Enforcer) IsFiltered() bool {
	filteredAdapter, ok := e.adapter.(persist.FilteredAdapter)
	if !ok {
		return false
	}
	return filteredAdapter.IsFiltered()
}

// SavePolicy saves the current policy (usually after changed with Casbin API) back to file/database.
func (e *Enforcer) SavePolicy() error {
	if e.IsFiltered() {
		return errors.New("cannot save a filtered policy")
	}
	if err := e.adapter.SavePolicy(e.model); err != nil {
		return err
	}
	if e.watcher != nil {
		var err error
		if watcher, ok := e.watcher.(persist.WatcherEx); ok {
			err = watcher.UpdateForSavePolicy(e.model)
		} else {
			err = e.watcher.Update()
		}
		return err
	}
	return nil
}

func (e *Enforcer) initRmMap() {
	for ptype, assertion := range e.model["g"] {
		if rm, ok := e.rmMap[ptype]; ok {
			_ = rm.Clear()
			continue
		}
		if len(assertion.Tokens) <= 2 && len(assertion.ParamsTokens) == 0 {
			assertion.RM = defaultrolemanager.NewRoleManagerImpl(10)
			e.rmMap[ptype] = assertion.RM
		}
		if len(assertion.Tokens) <= 2 && len(assertion.ParamsTokens) != 0 {
			assertion.CondRM = defaultrolemanager.NewConditionalRoleManager(10)
			e.condRmMap[ptype] = assertion.CondRM
		}
		if len(assertion.Tokens) > 2 {
			if len(assertion.ParamsTokens) == 0 {
				assertion.RM = defaultrolemanager.NewRoleManager(10)
				e.rmMap[ptype] = assertion.RM
			} else {
				assertion.CondRM = defaultrolemanager.NewConditionalDomainManager(10)
				e.condRmMap[ptype] = assertion.CondRM
			}
			matchFun := "keyMatch(r_dom, p_dom)"
			if strings.Contains(e.model["m"]["m"].Value, matchFun) {
				e.AddNamedDomainMatchingFunc(ptype, "g", util.KeyMatch)
			}
		}
	}
}

// EnableEnforce changes the enforcing state of Casbin, when Casbin is disabled, all access will be allowed by the Enforce() function.
func (e *Enforcer) EnableEnforce(enable bool) {
	e.enabled = enable
}

// EnableLog changes whether Casbin will log messages to the Logger.
func (e *Enforcer) EnableLog(enable bool) {
	e.logger.EnableLog(enable)
}

// IsLogEnabled returns the current logger's enabled status.
func (e *Enforcer) IsLogEnabled() bool {
	return e.logger.IsEnabled()
}

// EnableAutoNotifyWatcher controls whether to save a policy rule automatically notify the Watcher when it is added or removed.
func (e *Enforcer) EnableAutoNotifyWatcher(enable bool) {
	e.autoNotifyWatcher = enable
}

// EnableAutoNotifyDispatcher controls whether to save a policy rule automatically notify the Dispatcher when it is added or removed.
func (e *Enforcer) EnableAutoNotifyDispatcher(enable bool) {
	e.autoNotifyDispatcher = enable
}

// EnableAutoSave controls whether to save a policy rule automatically to the adapter when it is added or removed.
func (e *Enforcer) EnableAutoSave(autoSave bool) {
	e.autoSave = autoSave
}

// EnableAutoBuildRoleLinks controls whether to rebuild the role inheritance relations when a role is added or deleted.
func (e *Enforcer) EnableAutoBuildRoleLinks(autoBuildRoleLinks bool) {
	e.autoBuildRoleLinks = autoBuildRoleLinks
}

// EnableAcceptJsonRequest controls whether to accept json as a request parameter.
func (e *Enforcer) EnableAcceptJsonRequest(acceptJsonRequest bool) {
	e.acceptJsonRequest = acceptJsonRequest
}

// BuildRoleLinks manually rebuild the role inheritance relations.
func (e *Enforcer) BuildRoleLinks() error {
	if e.rmMap == nil {
		return errors.New("rmMap is nil")
	}
	for _, rm := range e.rmMap {
		err := rm.Clear()
		if err != nil {
			return err
		}
	}

	return e.model.BuildRoleLinks(e.rmMap)
}

// BuildIncrementalRoleLinks provides incremental build the role inheritance relations.
func (e *Enforcer) BuildIncrementalRoleLinks(op model.PolicyOp, ptype string, rules [][]string) error {
	e.invalidateMatcherMap()
	return e.model.BuildIncrementalRoleLinks(e.rmMap, op, "g", ptype, rules)
}

// BuildIncrementalConditionalRoleLinks provides incremental build the role inheritance relations with conditions.
func (e *Enforcer) BuildIncrementalConditionalRoleLinks(op model.PolicyOp, ptype string, rules [][]string) error {
	e.invalidateMatcherMap()
	return e.model.BuildIncrementalConditionalRoleLinks(e.condRmMap, op, "g", ptype, rules)
}

// NewEnforceContext Create a default structure based on the suffix.
func NewEnforceContext(suffix string) EnforceContext {
	return EnforceContext{
		RType: "r" + suffix,
		PType: "p" + suffix,
		EType: "e" + suffix,
		MType: "m" + suffix,
	}
}

func (e *Enforcer) invalidateMatcherMap() {
	e.matcherMap = sync.Map{}
}

// validateEnforcementState validates the enforcer's operational state before processing.
func (e *Enforcer) validateEnforcementState() error {
	if !e.enabled {
		return nil // Early return for disabled enforcer, not an error
	}

	if e.model == nil {
		return &EnforcementError{
			Type:    ErrConfigurationError,
			Message: "model is not initialized",
			Context: map[string]interface{}{"enabled": e.enabled},
		}
	}

	// Validate required model sections exist
	requiredSections := []string{"r", "p", "e", "m"}
	for _, section := range requiredSections {
		if _, exists := e.model[section]; !exists {
			return &EnforcementError{
				Type:    ErrConfigurationError,
				Message: fmt.Sprintf("required model section '%s' is missing", section),
				Context: map[string]interface{}{"section": section},
			}
		}
	}

	return nil
}

// prepareEnforcementContext prepares the execution context for policy evaluation.
func (e *Enforcer) prepareEnforcementContext(rvals []interface{}) (*EnforcementContext, error) {
	context := &EnforcementContext{
		RType: "r",
		PType: "p",
		EType: "e",
		MType: "m",
		RVals: rvals,
	}

	// Extract EnforceContext if present
	if len(rvals) != 0 {
		if enforceCtx, ok := rvals[0].(EnforceContext); ok {
			context.RType = enforceCtx.RType
			context.PType = enforceCtx.PType
			context.EType = enforceCtx.EType
			context.MType = enforceCtx.MType
			context.RVals = rvals[1:] // Remove EnforceContext from rvals
		}
	}

	// Validate that required types exist in model
	if _, exists := e.model["r"][context.RType]; !exists {
		return nil, &EnforcementError{
			Type:    ErrConfigurationError,
			Message: fmt.Sprintf("request type '%s' not found in model", context.RType),
			Context: map[string]interface{}{"rType": context.RType},
		}
	}

	if _, exists := e.model["p"][context.PType]; !exists {
		return nil, &EnforcementError{
			Type:    ErrConfigurationError,
			Message: fmt.Sprintf("policy type '%s' not found in model", context.PType),
			Context: map[string]interface{}{"pType": context.PType},
		}
	}

	// Build request and policy token mappings
	context.RTokens = make(map[string]int, len(e.model["r"][context.RType].Tokens))
	for i, token := range e.model["r"][context.RType].Tokens {
		context.RTokens[token] = i
	}

	context.PTokens = make(map[string]int, len(e.model["p"][context.PType].Tokens))
	for i, token := range e.model["p"][context.PType].Tokens {
		context.PTokens[token] = i
	}

	// Process JSON requests if enabled
	if e.acceptJsonRequest {
		if err := e.processJsonRequests(context); err != nil {
			return nil, err
		}
	}

	// Create enforcement parameters
	context.Parameters = enforceParameters{
		rTokens: context.RTokens,
		rVals:   context.RVals,
		pTokens: context.PTokens,
	}

	return context, nil
}

// processJsonRequests processes JSON request values if JSON support is enabled.
func (e *Enforcer) processJsonRequests(context *EnforcementContext) error {
	// Try to parse all request values from json to map[string]interface{}
	// Skip if there is an error
	for i, rval := range context.RVals {
		if rvalStr, ok := rval.(string); ok {
			if mapValue, err := util.JsonToMap(rvalStr); err == nil {
				context.RVals[i] = mapValue
			}
			// Note: We intentionally ignore JSON parsing errors as per original behavior
		}
	}
	return nil
}

// buildMatcherExpression compiles and caches matcher expressions with functions.
func (e *Enforcer) buildMatcherExpression(matcher string, context *EnforcementContext) (*govaluate.EvaluableExpression, error) {
	// Resolve matcher expression string
	if err := e.resolveMatcherExpression(matcher, context); err != nil {
		return nil, err
	}

	// Setup evaluation functions
	functions, err := e.setupEvaluationFunctions(context)
	if err != nil {
		return nil, err
	}

	// Setup eval function if needed
	if context.HasEval {
		functions["eval"] = generateEvalFunction(functions, &context.Parameters)
	}

	// Get or compile expression
	return e.getAndStoreMatcherExpression(context.HasEval, context.ExpString, functions)
}

// resolveMatcherExpression determines the matcher expression string to use.
func (e *Enforcer) resolveMatcherExpression(matcher string, context *EnforcementContext) error {
	if matcher == "" {
		// Use model matcher
		if _, exists := e.model["m"][context.MType]; !exists {
			return &EnforcementError{
				Type:    ErrConfigurationError,
				Message: fmt.Sprintf("matcher type '%s' not found in model", context.MType),
				Context: map[string]interface{}{"mType": context.MType},
			}
		}
		context.ExpString = e.model["m"][context.MType].Value
	} else {
		// Use custom matcher
		context.ExpString = util.RemoveComments(util.EscapeAssertion(matcher))
	}

	// Check if expression contains eval function
	context.HasEval = util.HasEval(context.ExpString)

	return nil
}

// setupEvaluationFunctions sets up the function map for expression evaluation.
func (e *Enforcer) setupEvaluationFunctions(context *EnforcementContext) (map[string]govaluate.ExpressionFunction, error) {
	functions := e.fm.GetFunctions()

	// Setup role manager functions
	if _, ok := e.model["g"]; ok {
		for key, ast := range e.model["g"] {
			// g must be a normal role definition (ast.RM != nil)
			// or a conditional role definition (ast.CondRM != nil)
			// ast.RM and ast.CondRM shouldn't be nil at the same time
			if ast.RM != nil {
				functions[key] = util.GenerateGFunction(ast.RM)
			}
			if ast.CondRM != nil {
				functions[key] = util.GenerateConditionalGFunction(ast.CondRM)
			}
		}
	}

	return functions, nil
}

// evaluatePolicies executes policy evaluation logic with proper separation of concerns.
func (e *Enforcer) evaluatePolicies(expression *govaluate.EvaluableExpression, context *EnforcementContext) (*PolicyEvaluationResult, error) {
	// Validate request format
	if err := e.validateRequestFormat(context); err != nil {
		return nil, err
	}

	policyLen := len(e.model["p"][context.PType].Policy)
	hasValidPolicies := policyLen != 0 && strings.Contains(context.ExpString, context.PType+"_")

	if hasValidPolicies {
		return e.evaluateAgainstPolicies(expression, context, policyLen)
	} else {
		return e.evaluateWithoutPolicies(expression, context)
	}
}

// validateRequestFormat validates the request format against model expectations.
func (e *Enforcer) validateRequestFormat(context *EnforcementContext) error {
	expectedLen := len(e.model["r"][context.RType].Tokens)
	actualLen := len(context.RVals)

	if expectedLen != actualLen {
		return &EnforcementError{
			Type:    ErrInvalidRequest,
			Message: fmt.Sprintf("invalid request size: expected %d, got %d", expectedLen, actualLen),
			Context: map[string]interface{}{
				"expected": expectedLen,
				"actual":   actualLen,
				"rvals":    context.RVals,
			},
		}
	}

	return nil
}

// evaluateAgainstPolicies evaluates request against all policies.
func (e *Enforcer) evaluateAgainstPolicies(expression *govaluate.EvaluableExpression, context *EnforcementContext, policyLen int) (*PolicyEvaluationResult, error) {
	result := &PolicyEvaluationResult{
		PolicyEffects: make([]effector.Effect, policyLen),
		MatchResults:  make([]float64, policyLen),
		ExplainIndex:  -1,
	}

	for policyIndex, pvals := range e.model["p"][context.PType].Policy {
		// Validate policy format
		if err := e.validatePolicyFormat(context, pvals); err != nil {
			return nil, err
		}

		// Evaluate single policy rule
		matchResult, err := e.evaluatePolicyRule(expression, context, pvals)
		if err != nil {
			return nil, err
		}
		result.MatchResults[policyIndex] = matchResult

		// Determine policy effect
		result.PolicyEffects[policyIndex] = e.determinePolicyEffect(context, pvals)

		// Merge effects and check for early termination
		effect, explainIndex, err := e.eft.MergeEffects(
			e.model["e"][context.EType].Value,
			result.PolicyEffects,
			result.MatchResults,
			policyIndex,
			policyLen,
		)
		if err != nil {
			return nil, err
		}

		result.Effect = effect
		result.ExplainIndex = explainIndex

		// Early termination if effect is determined
		if effect != effector.Indeterminate {
			break
		}
	}

	return result, nil
}

// evaluateWithoutPolicies evaluates request when no policies exist or don't apply.
func (e *Enforcer) evaluateWithoutPolicies(expression *govaluate.EvaluableExpression, context *EnforcementContext) (*PolicyEvaluationResult, error) {
	// Special case: eval() function requires policies
	if context.HasEval && len(e.model["p"][context.PType].Policy) == 0 {
		return nil, &EnforcementError{
			Type:    ErrEvaluationFailure,
			Message: "please make sure rule exists in policy when using eval() in matcher",
			Context: map[string]interface{}{"hasEval": context.HasEval},
		}
	}

	result := &PolicyEvaluationResult{
		PolicyEffects: make([]effector.Effect, 1),
		MatchResults:  make([]float64, 1),
		ExplainIndex:  -1,
	}
	result.MatchResults[0] = 1

	// Create empty policy values for evaluation
	context.Parameters.pVals = make([]string, len(context.Parameters.pTokens))

	// Evaluate expression
	evalResult, err := expression.Eval(context.Parameters)
	if err != nil {
		return nil, &EnforcementError{
			Type:    ErrEvaluationFailure,
			Message: fmt.Sprintf("expression evaluation failed: %v", err),
			Context: map[string]interface{}{"error": err.Error()},
		}
	}

	// Convert result to effect
	if boolResult, ok := evalResult.(bool); ok && boolResult {
		result.PolicyEffects[0] = effector.Allow
	} else {
		result.PolicyEffects[0] = effector.Indeterminate
	}

	// Merge effects
	effect, explainIndex, err := e.eft.MergeEffects(
		e.model["e"][context.EType].Value,
		result.PolicyEffects,
		result.MatchResults,
		0,
		1,
	)
	if err != nil {
		return nil, err
	}

	result.Effect = effect
	result.ExplainIndex = explainIndex

	return result, nil
}

// validatePolicyFormat validates a single policy rule format.
func (e *Enforcer) validatePolicyFormat(context *EnforcementContext, pvals []string) error {
	expectedLen := len(e.model["p"][context.PType].Tokens)
	actualLen := len(pvals)

	if expectedLen != actualLen {
		return &EnforcementError{
			Type:    ErrInvalidPolicy,
			Message: fmt.Sprintf("invalid policy size: expected %d, got %d", expectedLen, actualLen),
			Context: map[string]interface{}{
				"expected": expectedLen,
				"actual":   actualLen,
				"pvals":    pvals,
			},
		}
	}

	return nil
}

// evaluatePolicyRule evaluates a single policy rule against the request.
func (e *Enforcer) evaluatePolicyRule(expression *govaluate.EvaluableExpression, context *EnforcementContext, pvals []string) (float64, error) {
	context.Parameters.pVals = pvals

	evalResult, err := expression.Eval(context.Parameters)
	if err != nil {
		return 0, &EnforcementError{
			Type:    ErrEvaluationFailure,
			Message: fmt.Sprintf("policy rule evaluation failed: %v", err),
			Context: map[string]interface{}{
				"error": err.Error(),
				"pvals": pvals,
			},
		}
	}

	// Convert result to match value (0 or 1)
	switch result := evalResult.(type) {
	case bool:
		if result {
			return 1, nil
		}
		return 0, nil
	case float64:
		if result != 0 {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, &EnforcementError{
			Type:    ErrEvaluationFailure,
			Message: "matcher result should be bool, int or float",
			Context: map[string]interface{}{"result": result, "type": fmt.Sprintf("%T", result)},
		}
	}
}

// determinePolicyEffect determines the effect of a policy rule based on its configuration.
func (e *Enforcer) determinePolicyEffect(context *EnforcementContext, pvals []string) effector.Effect {
	// Check if policy has effect token
	if j, ok := context.Parameters.pTokens[context.PType+"_eft"]; ok {
		eft := pvals[j]
		switch eft {
		case "allow":
			return effector.Allow
		case "deny":
			return effector.Deny
		default:
			return effector.Indeterminate
		}
	}

	// Default effect is allow if no effect token is present
	return effector.Allow
}

// compileEnforcementResult compiles final enforcement result with explanations and logging.
func (e *Enforcer) compileEnforcementResult(evalResult *PolicyEvaluationResult, context *EnforcementContext, explains *[]string) (bool, error) {
	// Build explanations if requested
	logExplains := e.buildExplanations(evalResult, context, explains)

	// Convert effect to boolean result
	result := evalResult.Effect == effector.Allow

	// Log enforcement decision
	e.logger.LogEnforce(context.ExpString, context.RVals, result, logExplains)

	return result, nil
}

// buildExplanations constructs explanation data for enforcement decisions.
func (e *Enforcer) buildExplanations(evalResult *PolicyEvaluationResult, context *EnforcementContext, explains *[]string) [][]string {
	var logExplains [][]string

	if explains != nil {
		// Include existing explanations
		if len(*explains) > 0 {
			logExplains = append(logExplains, *explains)
		}

		// Add policy explanation if available
		if evalResult.ExplainIndex != -1 && len(e.model["p"][context.PType].Policy) > evalResult.ExplainIndex {
			*explains = e.model["p"][context.PType].Policy[evalResult.ExplainIndex]
			logExplains = append(logExplains, *explains)
		}
	}

	return logExplains
}

// enforce use a custom matcher to decides whether a "subject" can access a "object" with the operation "action", input parameters are usually: (matcher, sub, obj, act), use model matcher by default when matcher is "".
func (e *Enforcer) enforce(matcher string, explains *[]string, rvals ...interface{}) (ok bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v\n%s", r, debug.Stack())
		}
	}()

	// Phase 1: Validate enforcement state
	if validateErr := e.validateEnforcementState(); validateErr != nil {
		return false, validateErr
	}

	// Early return if enforcer is disabled
	if !e.enabled {
		return true, nil
	}

	// Phase 2: Prepare enforcement context
	context, err := e.prepareEnforcementContext(rvals)
	if err != nil {
		return false, err
	}

	// Phase 3: Build matcher expression
	expression, err := e.buildMatcherExpression(matcher, context)
	if err != nil {
		return false, err
	}

	// Phase 4: Evaluate policies
	evalResult, err := e.evaluatePolicies(expression, context)
	if err != nil {
		return false, err
	}

	// Phase 5: Compile enforcement result
	return e.compileEnforcementResult(evalResult, context, explains)
}

func (e *Enforcer) getAndStoreMatcherExpression(hasEval bool, expString string, functions map[string]govaluate.ExpressionFunction) (*govaluate.EvaluableExpression, error) {
	var expression *govaluate.EvaluableExpression
	var err error
	var cachedExpression, isPresent = e.matcherMap.Load(expString)

	if !hasEval && isPresent {
		expression = cachedExpression.(*govaluate.EvaluableExpression)
	} else {
		expression, err = govaluate.NewEvaluableExpressionWithFunctions(expString, functions)
		if err != nil {
			return nil, err
		}
		e.matcherMap.Store(expString, expression)
	}
	return expression, nil
}

// Enforce decides whether a "subject" can access a "object" with the operation "action", input parameters are usually: (sub, obj, act).
func (e *Enforcer) Enforce(rvals ...interface{}) (bool, error) {
	return e.enforce("", nil, rvals...)
}

// EnforceWithMatcher use a custom matcher to decides whether a "subject" can access a "object" with the operation "action", input parameters are usually: (matcher, sub, obj, act), use model matcher by default when matcher is "".
func (e *Enforcer) EnforceWithMatcher(matcher string, rvals ...interface{}) (bool, error) {
	return e.enforce(matcher, nil, rvals...)
}

// EnforceEx explain enforcement by informing matched rules.
func (e *Enforcer) EnforceEx(rvals ...interface{}) (bool, []string, error) {
	explain := []string{}
	result, err := e.enforce("", &explain, rvals...)
	return result, explain, err
}

// EnforceExWithMatcher use a custom matcher and explain enforcement by informing matched rules.
func (e *Enforcer) EnforceExWithMatcher(matcher string, rvals ...interface{}) (bool, []string, error) {
	explain := []string{}
	result, err := e.enforce(matcher, &explain, rvals...)
	return result, explain, err
}

// BatchEnforce enforce in batches.
func (e *Enforcer) BatchEnforce(requests [][]interface{}) ([]bool, error) {
	var results []bool
	for _, request := range requests {
		result, err := e.enforce("", nil, request...)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}
	return results, nil
}

// BatchEnforceWithMatcher enforce with matcher in batches.
func (e *Enforcer) BatchEnforceWithMatcher(matcher string, requests [][]interface{}) ([]bool, error) {
	var results []bool
	for _, request := range requests {
		result, err := e.enforce(matcher, nil, request...)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}
	return results, nil
}

// AddNamedMatchingFunc add MatchingFunc by ptype RoleManager.
func (e *Enforcer) AddNamedMatchingFunc(ptype, name string, fn rbac.MatchingFunc) bool {
	if rm, ok := e.rmMap[ptype]; ok {
		rm.AddMatchingFunc(name, fn)
		return true
	}
	return false
}

// AddNamedDomainMatchingFunc add MatchingFunc by ptype to RoleManager.
func (e *Enforcer) AddNamedDomainMatchingFunc(ptype, name string, fn rbac.MatchingFunc) bool {
	if rm, ok := e.rmMap[ptype]; ok {
		rm.AddDomainMatchingFunc(name, fn)
		return true
	}
	if condRm, ok := e.condRmMap[ptype]; ok {
		condRm.AddDomainMatchingFunc(name, fn)
		return true
	}
	return false
}

// AddNamedLinkConditionFunc Add condition function fn for Link userName->roleName,
// when fn returns true, Link is valid, otherwise invalid.
func (e *Enforcer) AddNamedLinkConditionFunc(ptype, user, role string, fn rbac.LinkConditionFunc) bool {
	if rm, ok := e.condRmMap[ptype]; ok {
		rm.AddLinkConditionFunc(user, role, fn)
		return true
	}
	return false
}

// AddNamedDomainLinkConditionFunc Add condition function fn for Link userName-> {roleName, domain},
// when fn returns true, Link is valid, otherwise invalid.
func (e *Enforcer) AddNamedDomainLinkConditionFunc(ptype, user, role string, domain string, fn rbac.LinkConditionFunc) bool {
	if rm, ok := e.condRmMap[ptype]; ok {
		rm.AddDomainLinkConditionFunc(user, role, domain, fn)
		return true
	}
	return false
}

// SetNamedLinkConditionFuncParams Sets the parameters of the condition function fn for Link userName->roleName.
func (e *Enforcer) SetNamedLinkConditionFuncParams(ptype, user, role string, params ...string) bool {
	if rm, ok := e.condRmMap[ptype]; ok {
		rm.SetLinkConditionFuncParams(user, role, params...)
		return true
	}
	return false
}

// SetNamedDomainLinkConditionFuncParams Sets the parameters of the condition function fn
// for Link userName->{roleName, domain}.
func (e *Enforcer) SetNamedDomainLinkConditionFuncParams(ptype, user, role, domain string, params ...string) bool {
	if rm, ok := e.condRmMap[ptype]; ok {
		rm.SetDomainLinkConditionFuncParams(user, role, domain, params...)
		return true
	}
	return false
}

// assumes bounds have already been checked.
type enforceParameters struct {
	rTokens map[string]int
	rVals   []interface{}

	pTokens map[string]int
	pVals   []string
}

// implements govaluate.Parameters.
func (p enforceParameters) Get(name string) (interface{}, error) {
	if name == "" {
		return nil, nil
	}

	switch name[0] {
	case 'p':
		i, ok := p.pTokens[name]
		if !ok {
			return nil, errors.New("No parameter '" + name + "' found.")
		}
		return p.pVals[i], nil
	case 'r':
		i, ok := p.rTokens[name]
		if !ok {
			return nil, errors.New("No parameter '" + name + "' found.")
		}
		return p.rVals[i], nil
	default:
		return nil, errors.New("No parameter '" + name + "' found.")
	}
}

func generateEvalFunction(functions map[string]govaluate.ExpressionFunction, parameters *enforceParameters) govaluate.ExpressionFunction {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("function eval(subrule string) expected %d arguments, but got %d", 1, len(args))
		}

		expression, ok := args[0].(string)
		if !ok {
			return nil, errors.New("argument of eval(subrule string) must be a string")
		}
		expression = util.EscapeAssertion(expression)
		expr, err := govaluate.NewEvaluableExpressionWithFunctions(expression, functions)
		if err != nil {
			return nil, fmt.Errorf("error while parsing eval parameter: %s, %s", expression, err.Error())
		}
		return expr.Eval(parameters)
	}
}
