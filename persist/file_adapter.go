package persist

import (
	"bufio"
	"bytes"
	"github.com/hsluoyz/casbin/util"
	"io"
	"os"
	"strings"
	"github.com/hsluoyz/casbin"
)

// The file adapter for policy persistence, can load policy from file or save policy to file.
type FileAdapter struct {
	filePath string
}

// The constructor for FileAdapter.
func NewFileAdapter(filePath string) *FileAdapter {
	a := FileAdapter{}
	a.filePath = filePath
	return &a
}

// Load policy from file.
func (a *FileAdapter) LoadPolicy(model casbin.Model) {
	if a.filePath == "" {
		return
	}

	err := a.loadPolicyFile(model, loadPolicyLine)
	if err != nil {
		panic(err)
	}
}

// Save policy to file.
func (a *FileAdapter) SavePolicy(model casbin.Model) {
	if a.filePath == "" {
		return
	}

	var tmp bytes.Buffer

	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			tmp.WriteString(ptype + ", ")
			tmp.WriteString(util.ArrayToString(rule))
			tmp.WriteString("\n")
		}
	}

	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			tmp.WriteString(ptype + ", ")
			tmp.WriteString(util.ArrayToString(rule))
			tmp.WriteString("\n")
		}
	}

	err := a.savePolicyFile(strings.TrimRight(tmp.String(), "\n"))
	if err != nil {
		panic(err)
	}
}

func (a *FileAdapter) loadPolicyFile(model casbin.Model, handler func(string, casbin.Model)) error {
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

func (a *FileAdapter) savePolicyFile(text string) error {
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
