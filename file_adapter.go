package casbin

import (
	"bufio"
	"bytes"
	"github.com/hsluoyz/casbin/util"
	"io"
	"os"
	"strings"
)

type fileAdapter struct {
	filePath string
}

func newFileAdapter(filePath string) *fileAdapter {
	a := fileAdapter{}
	a.filePath = filePath
	return &a
}

func (a *fileAdapter) loadPolicy(model Model) {
	a.loadPolicyFile(model, loadPolicyLine)
}

func (a *fileAdapter) savePolicy(model Model) {
	var tmp bytes.Buffer

	for ptype, ast := range model["p"] {
		for _, rule := range ast.policy {
			tmp.WriteString(ptype + ", ")
			tmp.WriteString(util.ArrayToString(rule))
			tmp.WriteString("\n")
		}
	}

	for ptype, ast := range model["g"] {
		for _, rule := range ast.policy {
			tmp.WriteString(ptype + ", ")
			tmp.WriteString(util.ArrayToString(rule))
			tmp.WriteString("\n")
		}
	}

	a.savePolicyFile(strings.TrimRight(tmp.String(), "\n"))
}

func (a *fileAdapter) loadPolicyFile(model Model, handler func(string, Model)) error {
	f, err := os.Open(a.filePath)
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

func (a *fileAdapter) savePolicyFile(text string) error {
	f, err := os.Create(a.filePath)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	w.WriteString(text)
	w.Flush()
	f.Close()
	return nil
}
