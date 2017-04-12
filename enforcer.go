package casbin

import (
	"github.com/Knetic/govaluate"
	"strings"
	"log"
)

type Enforcer struct {
	modelPath string
	policyPath string
	model   Model
}

func (enforcer *Enforcer) init(modelPath string, policyPath string) {
	enforcer.modelPath = modelPath
	enforcer.policyPath = policyPath

	enforcer.reload()
}

func (enforcer *Enforcer) reload() {
	enforcer.model = loadModel(enforcer.modelPath)
	printModel(enforcer.model)

	loadPolicy(enforcer.policyPath, enforcer.model)
}

func (enforcer *Enforcer) keyMatch(key1 string, key2 string) bool {
	i := strings.Index(key2, "*")
	if i == -1 {
		return key1 == key2
	} else {
		if len(key1) > i {
			return key1[:i] == key2[:i]
		} else {
			return key1 == key2[:i]
		}
	}
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
