package casbin

import (
	"log"
	"os"
	"bufio"
	"strings"
	"io"
	"bytes"
)

func loadPolicy(path string, model Model) {
	clearPolicy(model)
	log.Print("Policy:")
	loadPolicyFile(path, model, loadPolicyLine)
}

func savePolicy(path string, model Model) {
	var tmp bytes.Buffer

	for ptype, ast := range model["p"] {
		for _, rule := range ast.policy {
			tmp.WriteString(ptype + ", ")
			tmp.WriteString(arrayToString(rule))
			tmp.WriteString("\n")
		}
	}

	for ptype, ast := range model["g"] {
		for _, rule := range ast.policy {
			tmp.WriteString(ptype + ", ")
			tmp.WriteString(arrayToString(rule))
			tmp.WriteString("\n")
		}
	}

	savePolicyFile(path, strings.TrimRight(tmp.String(), "\n"))
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
