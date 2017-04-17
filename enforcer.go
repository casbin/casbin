package casbin

import (
	"github.com/Knetic/govaluate"
	"log"
)

// Enforcer is the main interface for authorization enforcement and policy management.
type Enforcer struct {
	modelPath  string
	policyPath string

	model      Model
	fm         FunctionMap

	enabled    bool
}

// Initialize an enforcer with a model file and a policy file.
func (enforcer *Enforcer) Init(modelPath string, policyPath string) {
	enforcer.modelPath = modelPath
	enforcer.policyPath = policyPath
	enforcer.enabled = true

	enforcer.LoadAll()
}

// Reload the model file and policy file, usually used when those files have been changed.
func (enforcer *Enforcer) LoadAll() {
	enforcer.model = loadModel(enforcer.modelPath)
	printModel(enforcer.model)
	enforcer.fm = loadFunctionMap()

	enforcer.LoadPolicy()
}

// Reload the policy file only.
func (enforcer *Enforcer) LoadPolicy() {
	loadPolicy(enforcer.policyPath, enforcer.model)
	printPolicy(enforcer.model)
	buildRoleLinks(enforcer.model)
}

// Save the current policy (usually changed with casbin API) back to the policy file.
func (enforcer *Enforcer) SavePolicy() {
	savePolicy(enforcer.policyPath, enforcer.model)
}

// Change the enforcing state of casbin, when casbin is disabled, all access will be allowed by the Enforce() function.
func (enforcer *Enforcer) Enable(enable bool) {
	enforcer.enabled = enable
}

// Decide whether a "subject" can access a "object" with the operation "action", input parameters are usually: (sub, obj, act).
func (enforcer *Enforcer) Enforce(rvals ...string) bool {
	if !enforcer.enabled {
		return true
	}

	expString := enforcer.model["m"]["m"].value
	var expression *govaluate.EvaluableExpression

	functions := make(map[string]govaluate.ExpressionFunction)

	for key, function := range enforcer.fm {
		functions[key] = function
	}

	_, ok := enforcer.model["g"]
	if ok {
		for key, ast := range enforcer.model["g"] {
			rm := ast.rm
			functions[key] = func(args ...interface{}) (interface{}, error) {
				name1 := args[0].(string)
				name2 := args[1].(string)

				return (bool)(rm.hasLink(name1, name2)), nil
			}
		}
	}
	expression, _ = govaluate.NewEvaluableExpressionWithFunctions(expString, functions)

	var policyResults []bool
	if len(enforcer.model["p"]["p"].policy) != 0 {
		policyResults = make([]bool, len(enforcer.model["p"]["p"].policy))

		for i, pvals := range enforcer.model["p"]["p"].policy {
			//log.Print("Policy Rule: ", pvals)

			parameters := make(map[string]interface{}, 8)
			for j, token := range enforcer.model["r"]["r"].tokens {
				parameters[token] = rvals[j]
			}
			for j, token := range enforcer.model["p"]["p"].tokens {
				parameters[token] = pvals[j]
			}

			result, _ := expression.Evaluate(parameters)
			//log.Print("Result: ", result)

			policyResults[i] = result.(bool)
		}
	} else {
		policyResults = make([]bool, 1)

		parameters := make(map[string]interface{}, 8)
		for j, token := range enforcer.model["r"]["r"].tokens {
			parameters[token] = rvals[j]
		}

		result, _ := expression.Evaluate(parameters)
		//log.Print("Result: ", result)

		policyResults[0] = result.(bool)
	}

	//log.Print("Rule Results: ", policyResults)

	result := false
	if enforcer.model["e"]["e"].value == "some(where (p_eft == allow))" {
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
func (enforcer *Enforcer) GetRoles(name string) []string {
	return enforcer.GetRolesForPolicyType("g", name)
}

// Get the roles assigned to a subject, policy type can be specified.
func (enforcer *Enforcer) GetRolesForPolicyType(ptype string, name string) []string {
	return enforcer.model["g"][ptype].rm.getRoles(name)
}

// Get the list of subjects that show up in the current policy.
func (enforcer *Enforcer) GetAllSubjects() []string {
	return getValuesForFieldInPolicy(enforcer.model, "p", "p", 0)
}

// Get the list of objects that show up in the current policy.
func (enforcer *Enforcer) GetAllObjects() []string {
	return getValuesForFieldInPolicy(enforcer.model, "p", "p", 1)
}

// Get the list of actions that show up in the current policy.
func (enforcer *Enforcer) GetAllActions() []string {
	return getValuesForFieldInPolicy(enforcer.model, "p", "p", 2)
}

// Get the list of roles that show up in the current policy.
func (enforcer *Enforcer) GetAllRoles() []string {
	return getValuesForFieldInPolicy(enforcer.model, "g", "g", 1)
}

// Get all the authorization rules in the policy.
func (enforcer *Enforcer) GetPolicy() [][]string {
	return enforcer.GetPolicyForPolicyType("p")
}

// Get all the authorization rules in the policy, policy type can be specified.
func (enforcer *Enforcer) GetPolicyForPolicyType(ptype string) [][]string {
	return getPolicy(enforcer.model, "p", ptype)
}

// Get all the authorization rules in the policy, a field filter can be specified.
func (enforcer *Enforcer) GetFilteredPolicy(fieldIndex int, fieldValue string) [][]string {
	return enforcer.GetFilteredPolicyForPolicyType("p", fieldIndex, fieldValue)
}

// Get all the authorization rules in the policy, a field filter can be specified, policy type can be specified.
func (enforcer *Enforcer) GetFilteredPolicyForPolicyType(ptype string, fieldIndex int, fieldValue string) [][]string {
	return getFilteredPolicy(enforcer.model, "p", ptype, fieldIndex, fieldValue)
}

// Get all the role inheritance rules in the policy.
func (enforcer *Enforcer) GetGroupingPolicy() [][]string {
	return enforcer.GetGroupingPolicyForPolicyType("g")
}

// Get all the role inheritance rules in the policy, policy type can be specified.
func (enforcer *Enforcer) GetGroupingPolicyForPolicyType(ptype string) [][]string {
	return getPolicy(enforcer.model, "g", ptype)
}

// Add an authorization rule to the current policy.
func (enforcer *Enforcer) AddPolicy(policy []string) {
	enforcer.AddPolicyForPolicyType("p", policy)
}

// Remove an authorization rule from the current policy.
func (enforcer *Enforcer) RemovePolicy(policy []string) {
	enforcer.RemovePolicyForPolicyType("p", policy)
}

// Add an authorization rule to the current policy, policy type can be specified.
func (enforcer *Enforcer) AddPolicyForPolicyType(ptype string, policy []string) {
	addPolicy(enforcer.model, "p", ptype, policy)
}

// Remove an authorization rule from the current policy, policy type can be specified.
func (enforcer *Enforcer) RemovePolicyForPolicyType(ptype string, policy []string) {
	removePolicy(enforcer.model, "p", ptype, policy)
}

// Add a role inheritance rule to the current policy.
func (enforcer *Enforcer) AddGroupingPolicy(policy []string) {
	enforcer.AddGroupingPolicyForPolicyType("g", policy)
	buildRoleLinks(enforcer.model)
}

// Remove a role inheritance rule from the current policy.
func (enforcer *Enforcer) RemoveGroupingPolicy(policy []string) {
	enforcer.RemoveGroupingPolicyForPolicyType("g", policy)
	buildRoleLinks(enforcer.model)
}

// Add a role inheritance rule to the current policy, policy type can be specified.
func (enforcer *Enforcer) AddGroupingPolicyForPolicyType(ptype string, policy []string) {
	addPolicy(enforcer.model, "g", ptype, policy)
	buildRoleLinks(enforcer.model)
}

// Remove a role inheritance rule from the current policy, policy type can be specified.
func (enforcer *Enforcer) RemoveGroupingPolicyForPolicyType(ptype string, policy []string) {
	removePolicy(enforcer.model, "g", ptype, policy)
	buildRoleLinks(enforcer.model)
}

// Add the function that gets attributes for a subject in ABAC.
func (enforcer *Enforcer) AddSubjectAttributeFunction(function Function) {
	addFunction(enforcer.fm, "subAttr", function)
}

// Add the function that gets attributes for a object in ABAC.
func (enforcer *Enforcer) AddObjectAttributeFunction(function Function) {
	addFunction(enforcer.fm, "objAttr", function)
}
