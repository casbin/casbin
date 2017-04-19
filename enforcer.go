package casbin

import (
	"github.com/Knetic/govaluate"
	"log"
)

// Enforcer is the main interface for authorization enforcement and policy management.
type Enforcer struct {
	modelPath string
	adapter   *FileAdapter

	model Model
	fm    FunctionMap

	enabled bool
}

// Initialize an enforcer with a model file and a policy file.
func (e *Enforcer) Init(modelPath string, policyPath string) {
	e.modelPath = modelPath
	e.adapter = NewFileAdapter(policyPath)

	e.enabled = true

	e.LoadModel()
	e.LoadPolicy()
}

// Reload the model from the model CONF file.
// Because the policy is attached to a model, so the policy is invalidated and needs to be loaded by yourself.
func (e *Enforcer) LoadModel() {
	e.model = loadModel(e.modelPath)
	printModel(e.model)
	e.fm = loadFunctionMap()
}

// Reload the policy.
func (e *Enforcer) LoadPolicy() {
	clearPolicy(e.model)
	e.adapter.LoadPolicy(e.model)

	log.Print("Policy:")
	printPolicy(e.model)

	buildRoleLinks(e.model)
}

// Save the current policy (usually changed with casbin API) back to the policy file.
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

				return (bool)(rm.hasLink(name1, name2)), nil
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

		result, _ := expression.Evaluate(parameters)
		//log.Print("Result: ", result)

		policyResults[0] = result.(bool)
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

// Get the roles assigned to a subject.
func (e *Enforcer) GetRoles(name string) []string {
	return e.GetRolesForPolicyType("g", name)
}

// Get the roles assigned to a subject, policy type can be specified.
func (e *Enforcer) GetRolesForPolicyType(ptype string, name string) []string {
	return e.model["g"][ptype].RM.getRoles(name)
}

// Get the list of subjects that show up in the current policy.
func (e *Enforcer) GetAllSubjects() []string {
	return getValuesForFieldInPolicy(e.model, "p", "p", 0)
}

// Get the list of objects that show up in the current policy.
func (e *Enforcer) GetAllObjects() []string {
	return getValuesForFieldInPolicy(e.model, "p", "p", 1)
}

// Get the list of actions that show up in the current policy.
func (e *Enforcer) GetAllActions() []string {
	return getValuesForFieldInPolicy(e.model, "p", "p", 2)
}

// Get the list of roles that show up in the current policy.
func (e *Enforcer) GetAllRoles() []string {
	return getValuesForFieldInPolicy(e.model, "g", "g", 1)
}

// Get all the authorization rules in the policy.
func (e *Enforcer) GetPolicy() [][]string {
	return e.GetPolicyForPolicyType("p")
}

// Get all the authorization rules in the policy, policy type can be specified.
func (e *Enforcer) GetPolicyForPolicyType(ptype string) [][]string {
	return getPolicy(e.model, "p", ptype)
}

// Get all the authorization rules in the policy, a field filter can be specified.
func (e *Enforcer) GetFilteredPolicy(fieldIndex int, fieldValue string) [][]string {
	return e.GetFilteredPolicyForPolicyType("p", fieldIndex, fieldValue)
}

// Get all the authorization rules in the policy, a field filter can be specified, policy type can be specified.
func (e *Enforcer) GetFilteredPolicyForPolicyType(ptype string, fieldIndex int, fieldValue string) [][]string {
	return getFilteredPolicy(e.model, "p", ptype, fieldIndex, fieldValue)
}

// Get all the role inheritance rules in the policy.
func (e *Enforcer) GetGroupingPolicy() [][]string {
	return e.GetGroupingPolicyForPolicyType("g")
}

// Get all the role inheritance rules in the policy, policy type can be specified.
func (e *Enforcer) GetGroupingPolicyForPolicyType(ptype string) [][]string {
	return getPolicy(e.model, "g", ptype)
}

// Add an authorization rule to the current policy.
func (e *Enforcer) AddPolicy(policy []string) {
	e.AddPolicyForPolicyType("p", policy)
}

// Remove an authorization rule from the current policy.
func (e *Enforcer) RemovePolicy(policy []string) {
	e.RemovePolicyForPolicyType("p", policy)
}

// Add an authorization rule to the current policy, policy type can be specified.
func (e *Enforcer) AddPolicyForPolicyType(ptype string, policy []string) {
	addPolicy(e.model, "p", ptype, policy)
}

// Remove an authorization rule from the current policy, policy type can be specified.
func (e *Enforcer) RemovePolicyForPolicyType(ptype string, policy []string) {
	removePolicy(e.model, "p", ptype, policy)
}

// Add a role inheritance rule to the current policy.
func (e *Enforcer) AddGroupingPolicy(policy []string) {
	e.AddGroupingPolicyForPolicyType("g", policy)
	buildRoleLinks(e.model)
}

// Remove a role inheritance rule from the current policy.
func (e *Enforcer) RemoveGroupingPolicy(policy []string) {
	e.RemoveGroupingPolicyForPolicyType("g", policy)
	buildRoleLinks(e.model)
}

// Add a role inheritance rule to the current policy, policy type can be specified.
func (e *Enforcer) AddGroupingPolicyForPolicyType(ptype string, policy []string) {
	addPolicy(e.model, "g", ptype, policy)
	buildRoleLinks(e.model)
}

// Remove a role inheritance rule from the current policy, policy type can be specified.
func (e *Enforcer) RemoveGroupingPolicyForPolicyType(ptype string, policy []string) {
	removePolicy(e.model, "g", ptype, policy)
	buildRoleLinks(e.model)
}

// Add the function that gets attributes for a subject in ABAC.
func (e *Enforcer) AddSubjectAttributeFunction(function Function) {
	addFunction(e.fm, "subAttr", function)
}

// Add the function that gets attributes for a object in ABAC.
func (e *Enforcer) AddObjectAttributeFunction(function Function) {
	addFunction(e.fm, "objAttr", function)
}
