package casbin

import (
	"github.com/Knetic/govaluate"
	"log"
)

type Enforcer struct {
	modelPath  string
	policyPath string
	model      Model
}

func (enforcer *Enforcer) init(modelPath string, policyPath string) {
	enforcer.modelPath = modelPath
	enforcer.policyPath = policyPath

	enforcer.loadAll()
}

func (enforcer *Enforcer) loadAll() {
	enforcer.model = loadModel(enforcer.modelPath)
	printModel(enforcer.model)

	enforcer.loadPolicy()
}

func (enforcer *Enforcer) loadPolicy() {
	loadPolicy(enforcer.policyPath, enforcer.model)
	printPolicy(enforcer.model)
	buildRoleLinks(enforcer.model)
}

func (enforcer *Enforcer) savePolicy() {
	savePolicy(enforcer.policyPath, enforcer.model)
}

func (enforcer *Enforcer) enforce(rvals ...string) bool {
	expString := enforcer.model["m"]["m"].value
	var expression *govaluate.EvaluableExpression = nil

	_, ok := enforcer.model["g"]
	if !ok {
		expression, _ = govaluate.NewEvaluableExpression(expString)
	} else {
		functions := make(map[string]govaluate.ExpressionFunction)

		for key, ast := range enforcer.model["g"] {
			functions[key] = func(args ...interface{}) (interface{}, error) {
				name1 := args[0].(string)
				name2 := args[1].(string)

				return (bool)(ast.rm.hasLink(name1, name2)), nil
			}
		}

		expression, _ = govaluate.NewEvaluableExpressionWithFunctions(expString, functions)
	}

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

func (enforcer *Enforcer) getRoles(name string) []string {
	return enforcer.getRolesForPolicyType("g", name)
}

func (enforcer *Enforcer) getRolesForPolicyType(ptype string, name string) []string {
	return enforcer.model["g"][ptype].rm.getRoles(name)
}

func (enforcer *Enforcer) getPolicy() [][]string {
	return enforcer.getPolicyForPolicyType("p")
}

func (enforcer *Enforcer) getPolicyForPolicyType(ptype string) [][]string {
	return getPolicy(enforcer.model, "p", ptype)
}

func (enforcer *Enforcer) getFilteredPolicy(fieldIndex int, fieldValue string) [][]string {
	return enforcer.getFilteredPolicyForPolicyType("p", fieldIndex, fieldValue)
}

func (enforcer *Enforcer) getFilteredPolicyForPolicyType(ptype string, fieldIndex int, fieldValue string) [][]string {
	return getFilteredPolicy(enforcer.model, "p", ptype, fieldIndex, fieldValue)
}

func (enforcer *Enforcer) getGroupingPolicy() [][]string {
	return enforcer.getGroupingPolicyForPolicyType("g")
}

func (enforcer *Enforcer) getGroupingPolicyForPolicyType(ptype string) [][]string {
	return getPolicy(enforcer.model, "g", ptype)
}

func (enforcer *Enforcer) addPolicy(policy []string) {
	enforcer.addPolicyForPolicyType("p", policy)
}

func (enforcer *Enforcer) removePolicy(policy []string) {
	enforcer.removePolicyForPolicyType("p", policy)
}

func (enforcer *Enforcer) addPolicyForPolicyType(ptype string, policy []string) {
	addPolicy(enforcer.model, "p", ptype, policy)
}

func (enforcer *Enforcer) removePolicyForPolicyType(ptype string, policy []string) {
	removePolicy(enforcer.model, "p", ptype, policy)
}

func (enforcer *Enforcer) addGroupingPolicy(policy []string) {
	enforcer.addGroupingPolicyForPolicyType("g", policy)
	buildRoleLinks(enforcer.model)
}

func (enforcer *Enforcer) removeGroupingPolicy(policy []string) {
	enforcer.removeGroupingPolicyForPolicyType("g", policy)
	buildRoleLinks(enforcer.model)
}

func (enforcer *Enforcer) addGroupingPolicyForPolicyType(ptype string, policy []string) {
	addPolicy(enforcer.model, "g", ptype, policy)
	buildRoleLinks(enforcer.model)
}

func (enforcer *Enforcer) removeGroupingPolicyForPolicyType(ptype string, policy []string) {
	removePolicy(enforcer.model, "g", ptype, policy)
	buildRoleLinks(enforcer.model)
}
