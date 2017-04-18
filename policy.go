package casbin

import (
	"log"
	"strings"
)

func buildRoleLinks(model Model) {
	for _, ast := range model["g"] {
		ast.buildRoleLinks()
	}
}

func printPolicy(model Model) {
	for key, ast := range model["p"] {
		log.Print(key, ": ", ast.value, ": ", ast.policy)
	}

	for key, ast := range model["g"] {
		log.Print(key, ": ", ast.value, ": ", ast.policy)
	}
}

func clearPolicy(model Model) {
	for _, ast := range model["p"] {
		ast.policy = nil
	}

	for _, ast := range model["g"] {
		ast.policy = nil
	}
}

func loadPolicyLine(line string, model Model) {
	if line == "" {
		return
	}

	tokens := strings.Split(line, ", ")

	key := tokens[0]
	sec := key[:1]
	model[sec][key].policy = append(model[sec][key].policy, tokens[1:])
}


func getPolicy(model Model, sec string, ptype string) [][]string {
	return model[sec][ptype].policy
}

func getFilteredPolicy(model Model, sec string, ptype string, fieldIndex int, fieldValue string) [][]string {
	res := [][]string{}

	for _, v := range model[sec][ptype].policy {
		if v[fieldIndex] == fieldValue {
			res = append(res, v)
		}
	}

	return res
}

func hasPolicy(model Model, sec string, ptype string, policy []string) bool {
	for _, rule := range model[sec][ptype].policy {
		if ArrayEquals(policy, rule) {
			return true
		}
	}

	return false
}

func addPolicy(model Model, sec string, ptype string, policy []string) bool {
	if !hasPolicy(model, sec, ptype, policy) {
		model[sec][ptype].policy = append(model[sec][ptype].policy, policy)
		return true
	} else {
		return false
	}
}

func removePolicy(model Model, sec string, ptype string, policy []string) bool {
	for i, rule := range model[sec][ptype].policy {
		if ArrayEquals(policy, rule) {
			model[sec][ptype].policy = append(model[sec][ptype].policy[:i], model[sec][ptype].policy[i+1:]...)
			return true
		}
	}

	return false
}

func getValuesForFieldInPolicy(model Model, sec string, ptype string, fieldIndex int) []string {
	users := []string{}

	for _, rule := range model[sec][ptype].policy {
		users = append(users, rule[fieldIndex])
	}

	ArrayRemoveDuplicates(&users)
	// sort.Strings(users)

	return users
}
