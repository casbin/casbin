package api

import (
	"github.com/Knetic/govaluate"
	"log"
	"github.com/hsluoyz/casbin"
	"github.com/hsluoyz/casbin/persist"
)

// Enforcer is the main interface for authorization enforcement and policy management.
type Enforcer struct {
	modelPath string
	model casbin.Model
	fm    casbin.FunctionMap

	adapter   persist.Adapter

	enabled bool
}

// Initialize an enforcer with a model file and a policy file.
func (e *Enforcer) InitWithFile(modelPath string, policyPath string) {
	e.modelPath = modelPath

	e.adapter = persist.NewFileAdapter(policyPath)

	e.enabled = true

	e.LoadModel()
	e.LoadPolicy()
}

// Initialize an enforcer with a model file and a policy from database.
func (e *Enforcer) InitWithDB(modelPath string, driverName string, dataSourceName string) {
	e.modelPath = modelPath

	e.adapter = persist.NewDBAdapter(driverName, dataSourceName)

	e.enabled = true

	e.LoadModel()
	e.LoadPolicy()
}

// Initialize an enforcer with a configuration file, by default is casbin.conf.
func (e *Enforcer) InitWithConfig(cfgPath string) {
	cfg := loadConfig(cfgPath)

	e.modelPath = cfg.modelPath

	if cfg.policyBackend == "file" {
		e.adapter = persist.NewFileAdapter(cfg.policyPath)
	} else if cfg.policyBackend == "database" {
		e.adapter = persist.NewDBAdapter(cfg.dbDriver, cfg.dbDataSource)
	}

	e.enabled = true

	e.LoadModel()
	e.LoadPolicy()
}

// Reload the model from the model CONF file.
// Because the policy is attached to a model, so the policy is invalidated and needs to be reloaded by calling LoadPolicy().
func (e *Enforcer) LoadModel() {
	e.model = casbin.LoadModel(e.modelPath)
	e.model.PrintModel()
	e.fm = casbin.LoadFunctionMap()
}

// Get the current model.
func (e *Enforcer) GetModel() casbin.Model {
	return e.model
}

// Clear all policy.
func (e *Enforcer) ClearPolicy() {
	e.model.ClearPolicy()
}

// Reload the policy from file/database.
func (e *Enforcer) LoadPolicy() {
	e.model.ClearPolicy()
	e.adapter.LoadPolicy(e.model)

	e.model.PrintPolicy()

	e.model.BuildRoleLinks()
}

// Save the current policy (usually after changed with casbin API) back to file/database.
func (e *Enforcer) SavePolicy() {
	e.adapter.SavePolicy(e.model)
}

// Change the enforcing state of casbin, when casbin is disabled, all access will be allowed by the Enforce() function.
func (e *Enforcer) Enable(enable bool) {
	e.enabled = enable
}

// Decide whether a "subject" can access a "object" with the operation "action", input parameters are usually: (sub, obj, act).
func (e *Enforcer) Enforce(rvals ...string) bool {
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
				name1 := args[0].(string)
				name2 := args[1].(string)

				return (bool)(rm.HasLink(name1, name2)), nil
			}
		}
	}
	expression, _ = govaluate.NewEvaluableExpressionWithFunctions(expString, functions)

	var policyResults []bool
	if len(e.model["p"]["p"].Policy) != 0 {
		policyResults = make([]bool, len(e.model["p"]["p"].Policy))

		for i, pvals := range e.model["p"]["p"].Policy {
			//log.Print("Policy Rule: ", pvals)

			parameters := make(map[string]interface{}, 8)
			for j, token := range e.model["r"]["r"].Tokens {
				parameters[token] = rvals[j]
			}
			for j, token := range e.model["p"]["p"].Tokens {
				parameters[token] = pvals[j]
			}

			result, _ := expression.Evaluate(parameters)
			//log.Print("Result: ", result)

			policyResults[i] = result.(bool)
		}
	} else {
		policyResults = make([]bool, 1)

		parameters := make(map[string]interface{}, 8)
		for j, token := range e.model["r"]["r"].Tokens {
			parameters[token] = rvals[j]
		}

		result, err := expression.Evaluate(parameters)
		//log.Print("Result: ", result)

		if err != nil {
			policyResults[0] = false
		} else {
			policyResults[0] = result.(bool)
		}
	}

	//log.Print("Rule Results: ", policyResults)

	result := false
	if e.model["e"]["e"].Value == "some(where (p_eft == allow))" {
		result = false
		for _, res := range policyResults {
			if res {
				result = true
				break
			}
		}
	}

	log.Print("Request ", rvals, ": ", result)

	return result
}
