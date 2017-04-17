package casbin

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"
	"strings"
)

func buildRoleLinks(model Model) {
	for _, ast := range model["g"] {
		ast.buildRoleLinks()
	}
}

func loadPolicy(path string, model Model) {
	clearPolicy(model)
	log.Print("Policy:")
	loadPolicyFile(path, model, loadPolicyLine)
}

func printPolicy(model Model) {
	for key, ast := range model["p"] {
		log.Print(key, ": ", ast.value, ": ", ast.policy)
	}

	for key, ast := range model["g"] {
		log.Print(key, ": ", ast.value, ": ", ast.policy)
	}
}

func arrayToString(s []string) string {
	var tmp bytes.Buffer
	for i, v := range s {
		if i != len(s)-1 {
			tmp.WriteString(v + ", ")
		} else {
			tmp.WriteString(v)
		}
	}
	return tmp.String()
}

func savePolicy(path string, model Model) {
	var tmp bytes.Buffer

	for key, ast := range model["p"] {
		for _, rule := range ast.policy {
			tmp.WriteString(key + ", ")
			tmp.WriteString(arrayToString(rule))
			tmp.WriteString("\n")
		}
	}

	for key, ast := range model["g"] {
		for _, rule := range ast.policy {
			tmp.WriteString(key + ", ")
			tmp.WriteString(arrayToString(rule))
			tmp.WriteString("\n")
		}
	}

	savePolicyFile(path, strings.TrimRight(tmp.String(), "\n"))
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

func loadPolicyFile(fileName string, model Model, handler func(string, Model)) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		handler(line, model)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

func savePolicyFile(fileName string, text string) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	w.WriteString(text)
	w.Flush()
	f.Close()
	return nil
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
		if arrayEquals(policy, rule) {
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
		if arrayEquals(policy, rule) {
			model[sec][ptype].policy = append(model[sec][ptype].policy[:i], model[sec][ptype].policy[i+1:]...)
			return true
		}
	}

	return false
}

func getValuesForFieldInPolicy(model Model, fieldIndex int) []string {
	users := []string{}

	for _, rule := range model["p"]["p"].policy {
		users = append(users, rule[fieldIndex])
	}

	arrayRemoveDuplicates(&users)
	// sort.Strings(users)

	return users
}
