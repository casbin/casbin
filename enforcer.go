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

	"github.com/Knetic/govaluate"
	"github.com/casbin/casbin/v2/effect"
	"github.com/casbin/casbin/v2/log"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	"github.com/casbin/casbin/v2/rbac"
	defaultrolemanager "github.com/casbin/casbin/v2/rbac/default-role-manager"
	"github.com/casbin/casbin/v2/util"
)

// Enforcer is the main interface for authorization enforcement and policy management.
type Enforcer struct {
	modelPath string
	model     model.Model
	fm        model.FunctionMap
	eft       effect.Effector

	adapter    persist.Adapter
	watcher    persist.Watcher
	dispatcher persist.Dispatcher
	rmMap      map[string]rbac.RoleManager

	enabled              bool
	autoSave             bool
	autoBuildRoleLinks   bool
	autoNotifyWatcher    bool
	autoNotifyDispatcher bool

	logger log.Logger
}

// NewEnforcer creates an enforcer via file or DB.
//
// File:
//
// 	e := casbin.NewEnforcer("path/to/basic_model.conf", "path/to/basic_policy.csv")
//
// MySQL DB:
//
// 	a := mysqladapter.NewDBAdapter("mysql", "mysql_username:mysql_password@tcp(127.0.0.1:3306)/")
// 	e := casbin.NewEnforcer("path/to/basic_model.conf", a)
//
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

	if paramLen-parsedParamLen == 2 {
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
	} else if paramLen-parsedParamLen == 1 {
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
	} else if paramLen-parsedParamLen == 0 {
		return e, nil
	} else {
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
}

func (e *Enforcer) initialize() {
	e.rmMap = map[string]rbac.RoleManager{}
	e.eft = effect.NewDefaultEffector()
	e.watcher = nil

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
	return watcher.SetUpdateCallback(func(string) { _ = e.LoadPolicy() })
}

// GetRoleManager gets the current role manager.
func (e *Enforcer) GetRoleManager() rbac.RoleManager {
	return e.rmMap["g"]
}

// SetRoleManager sets the current role manager.
func (e *Enforcer) SetRoleManager(rm rbac.RoleManager) {
	e.rmMap["g"] = rm
}

// SetEffector sets the current effector.
func (e *Enforcer) SetEffector(eft effect.Effector) {
	e.eft = eft
}

// ClearPolicy clears all policy.
func (e *Enforcer) ClearPolicy() {
	if e.dispatcher != nil && e.autoNotifyDispatcher {
		_ = e.dispatcher.ClearPolicy()
		return
	}
	e.model.ClearPolicy()
}

// LoadPolicy reloads the policy from file/database.
func (e *Enforcer) LoadPolicy() error {
	e.model.ClearPolicy()
	if err := e.adapter.LoadPolicy(e.model); err != nil && err.Error() != "invalid file path, file path cannot be empty" {
		return err
	}

	if err := e.model.SortPoliciesByPriority(); err != nil {
		return err
	}

	e.initRmMap()

	if e.autoBuildRoleLinks {
		err := e.BuildRoleLinks()
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Enforcer) loadFilteredPolicy(filter interface{}) error {
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
	for ptype := range e.model["g"] {
		e.rmMap[ptype] = defaultrolemanager.NewRoleManager(10)
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

// BuildRoleLinks manually rebuild the role inheritance relations.
func (e *Enforcer) BuildRoleLinks() error {
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
	return e.model.BuildIncrementalRoleLinks(e.rmMap, op, "g", ptype, rules)
}

// enforce use a custom matcher to decides whether a "subject" can access a "object" with the operation "action", input parameters are usually: (matcher, sub, obj, act), use model matcher by default when matcher is "".
func (e *Enforcer) enforce(matcher string, explains *[]string, rvals ...interface{}) (ok bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	if !e.enabled {
		return true, nil
	}

	functions := e.fm.GetFunctions()
	if _, ok := e.model["g"]; ok {
		for key, ast := range e.model["g"] {
			rm := ast.RM
			functions[key] = util.GenerateGFunction(rm)
		}
	}
	var expString string
	if matcher == "" {
		expString = e.model["m"]["m"].Value
	} else {
		expString = util.RemoveComments(util.EscapeAssertion(matcher))
	}

	var expression *govaluate.EvaluableExpression
	hasEval := util.HasEval(expString)

	if !hasEval {
		expression, err = govaluate.NewEvaluableExpressionWithFunctions(expString, functions)
		if err != nil {
			return false, err
		}
	}

	rTokens := make(map[string]int, len(e.model["r"]["r"].Tokens))
	for i, token := range e.model["r"]["r"].Tokens {
		rTokens[token] = i
	}
	pTokens := make(map[string]int, len(e.model["p"]["p"].Tokens))
	for i, token := range e.model["p"]["p"].Tokens {
		pTokens[token] = i
	}

	parameters := enforceParameters{
		rTokens: rTokens,
		rVals:   rvals,

		pTokens: pTokens,
	}

	var policyEffects []effect.Effect
	var matcherResults []float64

	if policyLen := len(e.model["p"]["p"].Policy); policyLen != 0 {
		policyEffects = make([]effect.Effect, policyLen)
		matcherResults = make([]float64, policyLen)
		if len(e.model["r"]["r"].Tokens) != len(rvals) {
			return false, fmt.Errorf(
				"invalid request size: expected %d, got %d, rvals: %v",
				len(e.model["r"]["r"].Tokens),
				len(rvals),
				rvals)
		}
		for i, pvals := range e.model["p"]["p"].Policy {
			// log.LogPrint("Policy Rule: ", pvals)
			if len(e.model["p"]["p"].Tokens) != len(pvals) {
				return false, fmt.Errorf(
					"invalid policy size: expected %d, got %d, pvals: %v",
					len(e.model["p"]["p"].Tokens),
					len(pvals),
					pvals)
			}

			parameters.pVals = pvals

			if hasEval {
				ruleNames := util.GetEvalValue(expString)
				replacements := make(map[string]string)
				for _, ruleName := range ruleNames {
					if j, ok := parameters.pTokens[ruleName]; ok {
						rule := util.EscapeAssertion(pvals[j])
						replacements[ruleName] = rule
					} else {
						return false, errors.New("please make sure rule exists in policy when using eval() in matcher")
					}
				}
				expWithRule := util.ReplaceEvalWithMap(expString, replacements)
				expression, err = govaluate.NewEvaluableExpressionWithFunctions(expWithRule, functions)
				if err != nil {
					return false, fmt.Errorf("p.sub_rule should satisfy the syntax of matcher: %s", err)
				}
			}

			result, err := expression.Eval(parameters)
			// log.LogPrint("Result: ", result)

			if err != nil {
				return false, err
			}

			switch result := result.(type) {
			case bool:
				if !result {
					policyEffects[i] = effect.Indeterminate
					continue
				}
			case float64:
				if result == 0 {
					policyEffects[i] = effect.Indeterminate
					continue
				} else {
					matcherResults[i] = result
				}
			default:
				return false, errors.New("matcher result should be bool, int or float")
			}

			if j, ok := parameters.pTokens["p_eft"]; ok {
				eft := parameters.pVals[j]
				if eft == "allow" {
					policyEffects[i] = effect.Allow
				} else if eft == "deny" {
					policyEffects[i] = effect.Deny
				} else {
					policyEffects[i] = effect.Indeterminate
				}
			} else {
				policyEffects[i] = effect.Allow
			}

			if e.model["e"]["e"].Value == "priority(p_eft) || deny" {
				break
			}

		}
	} else {
		if hasEval && len(e.model["p"]["p"].Policy) == 0 {
			return false, errors.New("please make sure rule exists in policy when using eval() in matcher")
		}

		policyEffects = make([]effect.Effect, 1)
		matcherResults = make([]float64, 1)

		parameters.pVals = make([]string, len(parameters.pTokens))

		result, err := expression.Eval(parameters)

		if err != nil {
			return false, err
		}

		if result.(bool) {
			policyEffects[0] = effect.Allow
		} else {
			policyEffects[0] = effect.Indeterminate
		}
	}

	result, explainIndex, err := e.eft.MergeEffects(e.model["e"]["e"].Value, policyEffects, matcherResults)
	if err != nil {
		return false, err
	}

	var logExplains [][]string

	if explains != nil {
		logExplains = append(logExplains, *explains)
		if explainIndex != -1 && len(e.model["p"]["p"].Policy) > explainIndex {
			*explains = e.model["p"]["p"].Policy[explainIndex]
		}
	}

	e.logger.LogEnforce(expString, rvals, result, logExplains)

	return result, nil
}

// Enforce decides whether a "subject" can access a "object" with the operation "action", input parameters are usually: (sub, obj, act).
func (e *Enforcer) Enforce(rvals ...interface{}) (bool, error) {
	return e.enforce("", nil, rvals...)
}

// EnforceWithMatcher use a custom matcher to decides whether a "subject" can access a "object" with the operation "action", input parameters are usually: (matcher, sub, obj, act), use model matcher by default when matcher is "".
func (e *Enforcer) EnforceWithMatcher(matcher string, rvals ...interface{}) (bool, error) {
	return e.enforce(matcher, nil, rvals...)
}

// EnforceEx explain enforcement by informing matched rules
func (e *Enforcer) EnforceEx(rvals ...interface{}) (bool, []string, error) {
	explain := []string{}
	result, err := e.enforce("", &explain, rvals...)
	return result, explain, err
}

// EnforceExWithMatcher use a custom matcher and explain enforcement by informing matched rules
func (e *Enforcer) EnforceExWithMatcher(matcher string, rvals ...interface{}) (bool, []string, error) {
	explain := []string{}
	result, err := e.enforce(matcher, &explain, rvals...)
	return result, explain, err
}

// BatchEnforce enforce in batches
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

// BatchEnforceWithMatcher enforce with matcher in batches
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

// AddNamedMatchingFunc add MatchingFunc by ptype RoleManager
func (e *Enforcer) AddNamedMatchingFunc(ptype, name string, fn defaultrolemanager.MatchingFunc) bool {
	if rm, ok := e.rmMap[ptype]; ok {
		rm.(*defaultrolemanager.RoleManager).AddMatchingFunc(name, fn)
		return true
	}
	return false
}

// AddNamedDomainMatchingFunc add MatchingFunc by ptype to RoleManager
func (e *Enforcer) AddNamedDomainMatchingFunc(ptype, name string, fn defaultrolemanager.MatchingFunc) bool {
	if rm, ok := e.rmMap[ptype]; ok {
		rm.(*defaultrolemanager.RoleManager).AddDomainMatchingFunc(name, fn)
		return true
	}
	return false
}

// assumes bounds have already been checked
type enforceParameters struct {
	rTokens map[string]int
	rVals   []interface{}

	pTokens map[string]int
	pVals   []string
}

// implements govaluate.Parameters
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
