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
	"fmt"
	"reflect"

	"github.com/Knetic/govaluate"
	"github.com/casbin/casbin/file-adapter"
	"github.com/casbin/casbin/model"
	"github.com/casbin/casbin/persist"
	"github.com/casbin/casbin/rbac"
	"github.com/casbin/casbin/rbac/default-role-manager"
	"github.com/casbin/casbin/util"
)

// Effect is the result for a policy rule.
type Effect int

// Values for policy effect.
const (
	EffectAllow Effect = iota
	EffectIndeterminate
	EffectDeny
)

// Enforcer is the main interface for authorization enforcement and policy management.
type Enforcer struct {
	modelPath string
	model     model.Model
	fm        model.FunctionMap
	rm        rbac.RoleManager

	adapter persist.Adapter
	watcher persist.Watcher

	enabled  bool
	autoSave bool
	autoBuildRoleLinks bool
}

// NewEnforcer creates an enforcer via file or DB.
// File:
// e := casbin.NewEnforcer("path/to/basic_model.conf", "path/to/basic_policy.conf")
// MySQL DB:
// a := mysqladapter.NewDBAdapter("mysql", "mysql_username:mysql_password@tcp(127.0.0.1:3306)/")
// e := casbin.NewEnforcer("path/to/basic_model.conf", a)
func NewEnforcer(params ...interface{}) *Enforcer {
	e := &Enforcer{}
	e.rm = defaultrolemanager.NewRoleManager(10)

	parsedParamLen := 0
	if len(params) >= 1 && reflect.TypeOf(params[len(params)-1]).Kind() == reflect.Bool {
		enableLog := params[len(params)-1].(bool)
		e.EnableLog(enableLog)

		parsedParamLen++
	}

	if len(params)-parsedParamLen == 2 {
		if reflect.TypeOf(params[0]).Kind() == reflect.String {
			if reflect.TypeOf(params[1]).Kind() == reflect.String {
				e.InitWithFile(params[0].(string), params[1].(string))
			} else {
				e.InitWithAdapter(params[0].(string), params[1].(persist.Adapter))
			}
		} else {
			if reflect.TypeOf(params[1]).Kind() == reflect.String {
				panic("Invalid parameters for enforcer.")
			} else {
				e.InitWithModelAndAdapter(params[0].(model.Model), params[1].(persist.Adapter))
			}
		}
	} else if len(params)-parsedParamLen == 1 {
		if reflect.TypeOf(params[0]).Kind() == reflect.String {
			e.InitWithFile(params[0].(string), "")
		} else {
			e.InitWithModelAndAdapter(params[0].(model.Model), nil)
		}
	} else if len(params)-parsedParamLen == 0 {
		e.InitWithFile("", "")
	} else {
		panic("Invalid parameters for enforcer.")
	}

	return e
}

// InitWithFile initializes an enforcer with a model file and a policy file.
func (e *Enforcer) InitWithFile(modelPath string, policyPath string) {
	e.modelPath = modelPath

	e.adapter = fileadapter.NewAdapter(policyPath)
	e.watcher = nil

	e.initialize()

	if e.modelPath != "" {
		e.LoadModel()
		e.LoadPolicy()
	}
}

// InitWithAdapter initializes an enforcer with a database adapter.
func (e *Enforcer) InitWithAdapter(modelPath string, adapter persist.Adapter) {
	e.modelPath = modelPath

	e.adapter = adapter
	e.watcher = nil

	e.initialize()

	if e.modelPath != "" {
		e.LoadModel()
		e.LoadPolicy()
	}
}

// InitWithModelAndAdapter initializes an enforcer with a model and a database adapter.
func (e *Enforcer) InitWithModelAndAdapter(m model.Model, adapter persist.Adapter) {
	e.modelPath = ""
	e.adapter = adapter
	e.watcher = nil

	e.model = m
	e.model.PrintModel()
	e.fm = model.LoadFunctionMap()

	e.initialize()

	if e.adapter != nil {
		e.LoadPolicy()
	}
}

func (e *Enforcer) initialize() {
	e.enabled = true
	e.autoSave = true
	e.autoBuildRoleLinks = true
}

// NewModel creates a model.
func NewModel(text ...string) model.Model {
	m := make(model.Model)

	if len(text) == 1 {
		m.LoadModelFromText(text[0])
	} else if len(text) != 0 {
		panic("Invalid parameters for model.")
	}

	return m
}

// LoadModel reloads the model from the model CONF file.
// Because the policy is attached to a model, so the policy is invalidated and needs to be reloaded by calling LoadPolicy().
func (e *Enforcer) LoadModel() {
	e.model = NewModel()
	e.model.LoadModel(e.modelPath)
	e.model.PrintModel()
	e.fm = model.LoadFunctionMap()
}

// GetModel gets the current model.
func (e *Enforcer) GetModel() model.Model {
	return e.model
}

// SetModel sets the current model.
func (e *Enforcer) SetModel(m model.Model) {
	e.model = m
	e.fm = model.LoadFunctionMap()
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
func (e *Enforcer) SetWatcher(watcher persist.Watcher) {
	e.watcher = watcher
	watcher.SetUpdateCallback(func (string) {e.LoadPolicy()})
}

// SetRoleManager sets the current role manager.
func (e *Enforcer) SetRoleManager(rm rbac.RoleManager) {
	e.rm = rm
}

// ClearPolicy clears all policy.
func (e *Enforcer) ClearPolicy() {
	e.model.ClearPolicy()
}

// LoadPolicy reloads the policy from file/database.
func (e *Enforcer) LoadPolicy() error {
	e.model.ClearPolicy()
	err := e.adapter.LoadPolicy(e.model)
	if err != nil {
		return err
	}

	e.model.PrintPolicy()
	if e.autoBuildRoleLinks {
		e.BuildRoleLinks()
	}
	return nil
}

// SavePolicy saves the current policy (usually after changed with Casbin API) back to file/database.
func (e *Enforcer) SavePolicy() error {
	err := e.adapter.SavePolicy(e.model)
	if err == nil {
		if e.watcher != nil {
			e.watcher.Update()
		}
	}
	return err
}

// EnableEnforce changes the enforcing state of Casbin, when Casbin is disabled, all access will be allowed by the Enforce() function.
func (e *Enforcer) EnableEnforce(enable bool) {
	e.enabled = enable
}

// EnableLog changes whether to print Casbin log to the standard output.
func (e *Enforcer) EnableLog(enable bool) {
	util.EnableLog = enable
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
func (e *Enforcer) BuildRoleLinks() {
	e.rm.Clear()
	e.model.BuildRoleLinks(e.rm)
}

// Enforce decides whether a "subject" can access a "object" with the operation "action", input parameters are usually: (sub, obj, act).
func (e *Enforcer) Enforce(rvals ...interface{}) bool {
	if !e.enabled {
		return true
	}

	expString := e.model["m"]["m"].Value
	var expression *govaluate.EvaluableExpression

	functions := make(map[string]govaluate.ExpressionFunction)

	for key, function := range e.fm {
		functions[key] = function
	}

	_, ok := e.model["g"]
	if ok {
		for key, ast := range e.model["g"] {
			rm := ast.RM
			functions[key] = func(args ...interface{}) (interface{}, error) {
				if rm == nil {
					name1 := args[0].(string)
					name2 := args[1].(string)

					return name1 == name2, nil
				}

				if len(args) == 2 {
					name1 := args[0].(string)
					name2 := args[1].(string)

					return (bool)(rm.HasLink(name1, name2)), nil
				}

				name1 := args[0].(string)
				name2 := args[1].(string)
				domain := args[2].(string)

				return (bool)(rm.HasLink(name1, name2, domain)), nil
			}
		}
	}
	expression, _ = govaluate.NewEvaluableExpressionWithFunctions(expString, functions)

	var policyResults []Effect
	if len(e.model["p"]["p"].Policy) != 0 {
		policyResults = make([]Effect, len(e.model["p"]["p"].Policy))

		for i, pvals := range e.model["p"]["p"].Policy {
			// util.LogPrint("Policy Rule: ", pvals)

			parameters := make(map[string]interface{}, 8)
			for j, token := range e.model["r"]["r"].Tokens {
				parameters[token] = rvals[j]
			}
			for j, token := range e.model["p"]["p"].Tokens {
				parameters[token] = pvals[j]
			}

			result, err := expression.Evaluate(parameters)
			// util.LogPrint("Result: ", result)

			if err != nil {
				policyResults[i] = EffectIndeterminate
				panic(err)
			} else {
				if !result.(bool) {
					policyResults[i] = EffectIndeterminate
				} else {
					if effect, ok := parameters["p_eft"]; ok {
						if effect == "allow" {
							policyResults[i] = EffectAllow
						} else if effect == "deny" {
							policyResults[i] = EffectDeny
						} else {
							policyResults[i] = EffectIndeterminate
						}
					} else {
						policyResults[i] = EffectAllow
					}

					if e.model["e"]["e"].Value == "priority(p_eft) || deny" {
						break
					}
				}
			}
		}
	} else {
		policyResults = make([]Effect, 1)

		parameters := make(map[string]interface{}, 8)
		for j, token := range e.model["r"]["r"].Tokens {
			parameters[token] = rvals[j]
		}
		for _, token := range e.model["p"]["p"].Tokens {
			parameters[token] = ""
		}

		result, err := expression.Evaluate(parameters)
		// util.LogPrint("Result: ", result)

		if err != nil {
			policyResults[0] = EffectIndeterminate
			panic(err)
		} else {
			if result.(bool) {
				policyResults[0] = EffectAllow
			} else {
				policyResults[0] = EffectIndeterminate
			}
		}
	}

	// util.LogPrint("Rule Results: ", policyResults)

	result := false
	if e.model["e"]["e"].Value == "some(where (p_eft == allow))" {
		result = false
		for _, eft := range policyResults {
			if eft == EffectAllow {
				result = true
				break
			}
		}
	} else if e.model["e"]["e"].Value == "!some(where (p_eft == deny))" {
		result = true
		for _, eft := range policyResults {
			if eft == EffectDeny {
				result = false
				break
			}
		}
	} else if e.model["e"]["e"].Value == "some(where (p_eft == allow)) && !some(where (p_eft == deny))" {
		result = false
		for _, eft := range policyResults {
			if eft == EffectAllow {
				result = true
			} else if eft == EffectDeny {
				result = false
				break
			}
		}
	} else if e.model["e"]["e"].Value == "priority(p_eft) || deny" {
		result = false
		for _, eft := range policyResults {
			if eft != EffectIndeterminate {
				if eft == EffectAllow {
					result = true
				} else {
					result = false
				}
				break
			}
		}
	}

	reqStr := "Request: "
	for i, rval := range rvals {
		if i != len(rvals)-1 {
			reqStr += fmt.Sprintf("%v, ", rval)
		} else {
			reqStr += fmt.Sprintf("%v", rval)
		}
	}
	reqStr += fmt.Sprintf(" ---> %t", result)
	util.LogPrint(reqStr)

	return result
}
