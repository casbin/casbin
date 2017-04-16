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
	enabled    bool
}

func (enforcer *Enforcer) Init(modelPath string, policyPath string) {
	enforcer.modelPath = modelPath
	enforcer.policyPath = policyPath
	enforcer.enabled = true

	enforcer.LoadAll()
}

func (enforcer *Enforcer) LoadAll() {
	enforcer.model = loadModel(enforcer.modelPath)
	printModel(enforcer.model)

	enforcer.LoadPolicy()
}

func (enforcer *Enforcer) LoadPolicy() {
	loadPolicy(enforcer.policyPath, enforcer.model)
	printPolicy(enforcer.model)
	buildRoleLinks(enforcer.model)
}

func (enforcer *Enforcer) SavePolicy() {
	savePolicy(enforcer.policyPath, enforcer.model)
}

func (enforcer *Enforcer) Enable(enable bool) {
	enforcer.enabled = enable
}

func (enforcer *Enforcer) Enforce(rvals ...string) bool {
	if !enforcer.enabled {
		return true
	}

	expString := enforcer.model["m"]["m"].value
	var expression *govaluate.EvaluableExpression

	functions := make(map[string]govaluate.ExpressionFunction)

	functions["keyMatch"] = func(args ...interface{}) (interface{}, error) {
		name1 := args[0].(string)
		name2 := args[1].(string)

		return (bool)(keyMatch(name1, name2)), nil
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

	policyResults := make([]bool, len(enforcer.model["p"]["p"].policy))

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

func (enforcer *Enforcer) GetRoles(name string) []string {
	return enforcer.GetRolesForPolicyType("g", name)
}

func (enforcer *Enforcer) GetRolesForPolicyType(ptype string, name string) []string {
	return enforcer.model["g"][ptype].rm.getRoles(name)
}

func (enforcer *Enforcer) GetPolicy() [][]string {
	return enforcer.GetPolicyForPolicyType("p")
}

func (enforcer *Enforcer) GetPolicyForPolicyType(ptype string) [][]string {
	return getPolicy(enforcer.model, "p", ptype)
}

func (enforcer *Enforcer) GetFilteredPolicy(fieldIndex int, fieldValue string) [][]string {
	return enforcer.GetFilteredPolicyForPolicyType("p", fieldIndex, fieldValue)
}

func (enforcer *Enforcer) GetFilteredPolicyForPolicyType(ptype string, fieldIndex int, fieldValue string) [][]string {
	return getFilteredPolicy(enforcer.model, "p", ptype, fieldIndex, fieldValue)
}

func (enforcer *Enforcer) GetGroupingPolicy() [][]string {
	return enforcer.GetGroupingPolicyForPolicyType("g")
}

func (enforcer *Enforcer) GetGroupingPolicyForPolicyType(ptype string) [][]string {
	return getPolicy(enforcer.model, "g", ptype)
}

func (enforcer *Enforcer) AddPolicy(policy []string) {
	enforcer.AddPolicyForPolicyType("p", policy)
}

func (enforcer *Enforcer) RemovePolicy(policy []string) {
	enforcer.RemovePolicyForPolicyType("p", policy)
}

func (enforcer *Enforcer) AddPolicyForPolicyType(ptype string, policy []string) {
	addPolicy(enforcer.model, "p", ptype, policy)
}

func (enforcer *Enforcer) RemovePolicyForPolicyType(ptype string, policy []string) {
	removePolicy(enforcer.model, "p", ptype, policy)
}

func (enforcer *Enforcer) AddGroupingPolicy(policy []string) {
	enforcer.AddGroupingPolicyForPolicyType("g", policy)
	buildRoleLinks(enforcer.model)
}

func (enforcer *Enforcer) RemoveGroupingPolicy(policy []string) {
	enforcer.RemoveGroupingPolicyForPolicyType("g", policy)
	buildRoleLinks(enforcer.model)
}

func (enforcer *Enforcer) AddGroupingPolicyForPolicyType(ptype string, policy []string) {
	addPolicy(enforcer.model, "g", ptype, policy)
	buildRoleLinks(enforcer.model)
}

func (enforcer *Enforcer) RemoveGroupingPolicyForPolicyType(ptype string, policy []string) {
	removePolicy(enforcer.model, "g", ptype, policy)
	buildRoleLinks(enforcer.model)
}
