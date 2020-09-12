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
	"strings"
	"sync/atomic"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/casbin/casbin/v3/effect"
	"github.com/casbin/casbin/v3/internal"
	"github.com/casbin/casbin/v3/log"
	"github.com/casbin/casbin/v3/model"
	"github.com/casbin/casbin/v3/persist"
	fileadapter "github.com/casbin/casbin/v3/persist/file-adapter"
	"github.com/casbin/casbin/v3/rbac"
	defaultrolemanager "github.com/casbin/casbin/v3/rbac/default-role-manager"
	"github.com/casbin/casbin/v3/util"
)

// Enforcer is the main interface for authorization enforcement and policy management.
type Enforcer struct {
	modelPath string
	model     *model.Model
	fm        model.FunctionMap
	eft       effect.Effector

	adapter    persist.Adapter
	watcher    persist.Watcher
	dispatcher persist.Dispatcher
	rm         rbac.RoleManager

	internal             internal.PolicyManager
	enabled              bool
	autoSave             bool
	autoBuildRoleLinks   bool
	autoNotifyWatcher    bool
	autoNotifyDispatcher bool

	stopAutoLoad    chan struct{}
	autoLoadRunning int32
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
	e := &Enforcer{}

	parsedParamLen := 0
	paramLen := len(params)
	if paramLen >= 1 {
		enableLog, ok := params[paramLen-1].(bool)
		if ok {
			e.EnableLog(enableLog)

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
				err := e.InitWithModelAndAdapter(p0.(*model.Model), params[1].(persist.Adapter))
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
			err := e.InitWithModelAndAdapter(p0.(*model.Model), nil)
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
func (e *Enforcer) InitWithModelAndAdapter(m *model.Model, adapter persist.Adapter) error {
	e.adapter = adapter

	e.model = m
	e.model.PrintModel()
	e.fm = model.LoadFunctionMap()
	e.initialize()
	e.internal = internal.NewPolicyManager(m, adapter, e.rm)
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

func (e *Enforcer) initialize() {
	e.rm = defaultrolemanager.NewRoleManager(10)
	e.eft = effect.NewDefaultEffector()
	e.watcher = nil

	e.enabled = true
	e.autoSave = true
	e.autoBuildRoleLinks = true
	e.autoNotifyWatcher = true
}

// LoadModel reloads the model from the model CONF file.
// Because the policy is attached to a model, so the policy is invalidated and needs to be reloaded by calling LoadPolicy().
func (e *Enforcer) LoadModel() error {
	var err error
	e.model, err = model.NewModelFromFile(e.modelPath)
	if err != nil {
		return err
	}

	e.model.PrintModel()
	e.fm = model.LoadFunctionMap()

	e.initialize()

	return nil
}

// GetModel gets the current model.
func (e *Enforcer) GetModel() *model.Model {
	return e.model
}

// SetModel sets the current model.
func (e *Enforcer) SetModel(m *model.Model) {
	e.model = m
	e.fm = model.LoadFunctionMap()
	e.internal = internal.NewPolicyManager(m, e.adapter, e.rm)
	e.initialize()
}

// GetAdapter gets the current adapter.
func (e *Enforcer) GetAdapter() persist.Adapter {
	return e.adapter
}

// SetAdapter sets the current adapter.
func (e *Enforcer) SetAdapter(adapter persist.Adapter) {
	e.adapter = adapter
	e.internal = internal.NewPolicyManager(e.model, adapter, e.rm)
}

// GetPolicyManager gets the current policy manager.
func (e *Enforcer) GetPolicyManager() internal.PolicyManager {
	return e.internal
}

// SetWatcher sets the current watcher.
func (e *Enforcer) SetWatcher(watcher persist.Watcher) error {
	e.watcher = watcher
	return watcher.SetUpdateCallback(func(string) { _ = e.LoadPolicy() })
}

// SetDispatcher sets the current dispatcher.
func (e *Enforcer) SetDispatcher(dispatcher persist.Dispatcher) error {
	e.dispatcher = dispatcher
	return dispatcher.SetEnforcer(e)
}

// GetRoleManager gets the current role manager.
func (e *Enforcer) GetRoleManager() rbac.RoleManager {
	return e.rm
}

// SetRoleManager sets the current role manager.
func (e *Enforcer) SetRoleManager(rm rbac.RoleManager) {
	e.rm = rm
	e.internal = internal.NewPolicyManager(e.model, e.adapter, rm)
}

// SetEffector sets the current effector.
func (e *Enforcer) SetEffector(eft effect.Effector) {
	e.eft = eft
}

func (e *Enforcer) IsAudoLoadRunning() bool {
	return atomic.LoadInt32(&(e.autoLoadRunning)) != 0
}

// StartAutoLoadPolicy starts a go routine that will every specified duration call LoadPolicy
func (e *Enforcer) StartAutoLoadPolicy(d time.Duration) {
	// Don't start another goroutine if there is already one running
	if e.IsAudoLoadRunning() {
		return
	}
	atomic.StoreInt32(&(e.autoLoadRunning), int32(1))
	ticker := time.NewTicker(d)
	go func() {
		defer func() {
			ticker.Stop()
			atomic.StoreInt32(&(e.autoLoadRunning), int32(0))
		}()
		n := 1
		log.LogPrintf("Start automatically load policy")
		for {
			select {
			case <-ticker.C:
				// error intentionally ignored
				_ = e.LoadPolicy()
				// Uncomment this line to see when the policy is loaded.
				// log.Print("Load policy for time: ", n)
				n++
			case <-e.stopAutoLoad:
				log.LogPrintf("Stop automatically load policy")
				return
			}
		}
	}()
}

// StopAutoLoadPolicy causes the go routine to exit.
func (e *Enforcer) StopAutoLoadPolicy() {
	if e.IsAudoLoadRunning() {
		e.stopAutoLoad <- struct{}{}
	}
}

// ClearPolicy clears all policy.
func (e *Enforcer) ClearPolicy() error {
	if e.dispatcher != nil && e.autoNotifyDispatcher {
		return e.dispatcher.ClearPolicy()
	}
	// TODO: implement ClearPolicy in adapter and move model.ClearPolicy after adapter.ClearPolicy
	e.model.ClearPolicy()

	if err := e.adapter.SavePolicy(e.model); err != nil {
		return err
	}

	if log.GetLogger().IsEnabled() {
		log.LogPrint("Policy Management, Clear all policy")
	}

	return nil
}

// LoadPolicy reloads the policy from file/database.
func (e *Enforcer) LoadPolicy() error {
	e.model.ClearPolicy()
	if err := e.adapter.LoadPolicy(e.model); err != nil && err.Error() != "invalid file path, file path cannot be empty" {
		return err
	}

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

	e.model.PrintPolicy()
	if e.autoBuildRoleLinks {
		err := e.BuildRoleLinks()
		if err != nil {
			return err
		}
	}
	return nil
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

// EnableEnforce changes the enforcing state of Casbin, when Casbin is disabled, all access will be allowed by the Enforce() function.
func (e *Enforcer) EnableEnforce(enable bool) {
	e.enabled = enable
}

// EnableLog changes whether Casbin will log messages to the Logger.
func (e *Enforcer) EnableLog(enable bool) {
	log.GetLogger().EnableLog(enable)
}

// EnableAutoNotifyWatcher controls whether to save a policy rule automatically notify the Watcher when it is added or removed.
func (e *Enforcer) EnableAutoNotifyWatcher(enable bool) {
	e.autoNotifyWatcher = enable
}

// EnableautoNotifyDispatcher controls whether to save a policy rule automatically notify the Dispatcher when it is added or removed.
func (e *Enforcer) EnableautoNotifyDispatcher(enable bool) {
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
	err := e.rm.Clear()
	if err != nil {
		return err
	}

	return e.model.BuildRoleLinks(e.rm)
}

// BuildIncrementalRoleLinks provides incremental build the role inheritance relations.
func (e *Enforcer) BuildIncrementalRoleLinks(op model.PolicyOp, ptype string, rules [][]string) error {
	return e.model.BuildIncrementalRoleLinks(e.rm, op, "g", ptype, rules)
}

// enforce use a custom matcher to decides whether a "subject" can access a "object" with the operation "action", input parameters are usually: (matcher, sub, obj, act), use model matcher by default when matcher is "".
func (e *Enforcer) enforce(matcher string, explains *[][]string, rvals ...interface{}) (ok bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	if !e.enabled {
		return true, nil
	}

	functions := e.model.GenerateFunctions(e.fm)
	var expString string
	if matcher == "" {
		expString = e.model.GetMatcher()
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

	rTokens := e.model.GetTokens("r", "r")
	pTokens := e.model.GetTokens("p", "p")

	parameters := enforceParameters{
		rTokens: rTokens,
		rVals:   rvals,

		pTokens: pTokens,
	}
	policies := e.model.GetPolicy("p", "p")
	policyLen := len(policies)
	var cap int
	if policyLen > 0 {
		cap = policyLen
	} else {
		cap = 1
	}
	eftStream := e.eft.NewStream(e.model.GetEffectExpression(), cap)
	if policyLen != 0 {
		if len(parameters.rTokens) != len(rvals) {
			return false, fmt.Errorf(
				"invalid request size: expected %d, got %d, rvals: %v",
				len(parameters.rTokens),
				len(rvals),
				rvals)
		}
		for _, pvals := range policies {
			// log.LogPrint("Policy Rule: ", pvals)
			if len(parameters.pTokens) != len(pvals) {
				return false, fmt.Errorf(
					"invalid policy size: expected %d, got %d, pvals: %v",
					len(parameters.pTokens),
					len(pvals),
					pvals)
			}

			parameters.pVals = pvals

			if hasEval {
				ruleNames := util.GetEvalValue(expString)
				var expWithRule = expString
				for _, ruleName := range ruleNames {
					if j, ok := parameters.pTokens[ruleName]; ok {
						rule := util.EscapeAssertion(pvals[j])
						expWithRule = util.ReplaceEval(expWithRule, rule)
					} else {
						return false, errors.New("please make sure rule exists in policy when using eval() in matcher")
					}

					expression, err = govaluate.NewEvaluableExpressionWithFunctions(expWithRule, functions)
					if err != nil {
						return false, fmt.Errorf("p.sub_rule should satisfy the syntax of matcher: %s", err)
					}
				}

			}

			result, err := expression.Eval(parameters)
			// log.LogPrint("Result: ", result)

			if err != nil {
				return false, err
			}

			var eft effect.Effect
			if !result.(bool) {
				eft = effect.Indeterminate
			} else {
				eft = effect.Allow
			}

			if eft == effect.Indeterminate {
				eft = effect.Indeterminate
			} else if j, ok := parameters.pTokens["p_eft"]; ok {
				pEft := parameters.pVals[j]
				if pEft == "allow" {
					eft = effect.Allow
				} else if pEft == "deny" {
					eft = effect.Deny
				} else {
					eft = effect.Indeterminate
				}
			} else {
				eft = effect.Allow
			}

			if eftStream.PushEffect(eft) {
				break
			}
		}
	} else {

		parameters.pVals = make([]string, len(parameters.pTokens))

		result, err := expression.Eval(parameters)
		// log.LogPrint("Result: ", result)

		if err != nil {
			return false, err
		}

		var eft effect.Effect
		if result.(bool) {
			eft = effect.Allow
		} else {
			eft = effect.Indeterminate
		}

		eftStream.PushEffect(eft)
	}

	// log.LogPrint("Rule Results: ", policyEffects)

	result := eftStream.Next()
	explainIndexes := eftStream.Explain()

	if explains != nil {
		if explainIndexes != nil {
			var tempExpl [][]string = [][]string{}
			for _, index := range explainIndexes {
				// *explains = e.model.data["p"]["p"].Policy[explainIndex]
				tempExpl = append(tempExpl, policies[index])
			}
			*explains = tempExpl
		}
	}

	// Log request.
	if log.GetLogger().IsEnabled() {
		var reqStr strings.Builder
		reqStr.WriteString("Request: ")
		for i, rval := range rvals {
			if i != len(rvals)-1 {
				reqStr.WriteString(fmt.Sprintf("%v, ", rval))
			} else {
				reqStr.WriteString(fmt.Sprintf("%v", rval))
			}
		}
		reqStr.WriteString(fmt.Sprintf(" ---> %t\n", result))

		if explains != nil {
			reqStr.WriteString("Hit Policy: ")
			for _, policy := range *explains {
				for i, pval := range policy {
					if i != len(policy)-1 {
						reqStr.WriteString(fmt.Sprintf("%v, ", pval))
					} else {
						reqStr.WriteString(fmt.Sprintf("%v \n", pval))
					}
				}
			}

		}

		log.LogPrint(reqStr.String())
	}

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
func (e *Enforcer) EnforceEx(rvals ...interface{}) (bool, [][]string, error) {
	explain := [][]string{}
	result, err := e.enforce("", &explain, rvals...)
	return result, explain, err
}

// EnforceExWithMatcher use a custom matcher and explain enforcement by informing matched rules
func (e *Enforcer) EnforceExWithMatcher(matcher string, rvals ...interface{}) (bool, [][]string, error) {
	explain := [][]string{}
	result, err := e.enforce(matcher, &explain, rvals...)
	return result, explain, err
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
