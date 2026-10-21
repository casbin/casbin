// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package casbin

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

var updateConformance = flag.Bool("update-conformance", false, "regenerate the conformance cases from examples/")

const maxConformanceTests = 300

var conformanceSources = [][2]string{
	{"basic_model.conf", "basic_policy.csv"},
	{"basic_model.conf", "basic_inverse_policy.csv"},
	{"basic_model_without_spaces.conf", "basic_policy.csv"},
	{"basic_with_root_model.conf", "basic_policy.csv"},
	{"basic_without_resources_model.conf", "basic_without_resources_policy.csv"},
	{"basic_without_users_model.conf", "basic_without_users_policy.csv"},
	{"comment_model.conf", "basic_policy.csv"},
	{"eval_operator_model.conf", "eval_operator_policy.csv"},
	{"glob_model.conf", "glob_policy.csv"},
	{"ipmatch_model.conf", "ipmatch_policy.csv"},
	{"keymatch_model.conf", "keymatch_policy.csv"},
	{"keymatch2_model.conf", "keymatch2_policy.csv"},
	{"keymatch_with_rbac_in_domain.conf", "keymatch_with_rbac_in_domain.csv"},
	{"orbac_model.conf", "orbac_policy.csv"},
	{"priority_model.conf", "priority_policy.csv"},
	{"priority_model.conf", "priority_indeterminate_policy.csv"},
	{"priority_model_explicit.conf", "priority_policy_explicit.csv"},
	{"rbac_model.conf", "rbac_policy.csv"},
	{"rbac_model.conf", "rbac_with_hierarchy_policy.csv"},
	{"rbac_model_in_multi_line.conf", "rbac_policy.csv"},
	{"rbac_model_matcher_using_in_op.conf", "rbac_policy.csv"},
	{"rbac_model_matcher_using_in_op_bracket.conf", "rbac_policy.csv"},
	{"rbac_with_deny_model.conf", "rbac_with_deny_policy.csv"},
	{"rbac_with_not_deny_model.conf", "rbac_with_deny_policy.csv"},
	{"rbac_with_different_types_of_roles_model.conf", "rbac_with_different_types_of_roles_policy.csv"},
	{"rbac_with_domains_model.conf", "rbac_with_domains_policy.csv"},
	{"rbac_with_domains_model.conf", "rbac_with_domains_policy2.csv"},
	{"rbac_with_domains_model.conf", "rbac_with_hierarchy_with_domains_policy.csv"},
	{"rbac_with_multiple_policy_model.conf", "rbac_with_multiple_policy_policy.csv"},
	{"rbac_with_resource_roles_model.conf", "rbac_with_resource_roles_policy.csv"},
	{"rebac_model.conf", "rebac_policy.csv"},
	{"subject_priority_model.conf", "subject_priority_policy.csv"},
	{"subject_priority_model_with_domain.conf", "subject_priority_policy_with_domain.csv"},
}

type conformanceCase struct {
	Model  string            `json:"model"`
	Policy string            `json:"policy"`
	Tests  []conformanceTest `json:"tests"`
}

type conformanceTest struct {
	Request []string `json:"request"`
	Allow   bool     `json:"allow"`
	Explain []string `json:"explain"`
}

func conformanceCaseName(modelFile, policyFile string) string {
	m := strings.TrimSuffix(strings.TrimSuffix(modelFile, ".conf"), "_model")
	p := strings.TrimSuffix(strings.TrimSuffix(policyFile, ".csv"), "_policy")
	if m == p {
		return m
	}
	return m + "__" + p
}

var conformanceQuoted = regexp.MustCompile(`'([^']*)'|"([^"]*)"`)

var conformanceExtraValues = map[string][]string{
	"ipmatch_model.conf": {"192.168.2.1", "192.168.2.123", "192.168.0.1", "10.0.0.5", "127.0.0.1"},
}

func conformanceCandidates(e *Enforcer, extra []string) [][]string {
	m := e.GetModel()
	n := len(m["r"]["r"].Tokens)
	shared := map[string]bool{"unknown": true}
	for _, v := range extra {
		shared[v] = true
	}
	columns := make([]map[string]bool, n)
	for i := range columns {
		columns[i] = map[string]bool{}
	}
	rIndex := map[string]int{}
	for i, token := range m["r"]["r"].Tokens {
		rIndex[strings.TrimPrefix(token, "r_")] = i
	}
	for _, sec := range []string{"p", "g"} {
		for ptype, ast := range m[sec] {
			for _, rule := range ast.Policy {
				for i, v := range rule {
					if sec == "g" {
						shared[v] = true
					} else if i < len(ast.Tokens) {
						if j, ok := rIndex[strings.TrimPrefix(ast.Tokens[i], ptype+"_")]; ok {
							columns[j][v] = true
						} else if i < n {
							columns[i][v] = true
						}
					}
					for _, match := range conformanceQuoted.FindAllStringSubmatch(v, -1) {
						shared[match[1]+match[2]] = true
					}
				}
			}
		}
	}

	requests := [][]string{{}}
	for _, column := range columns {
		set := map[string]bool{}
		for v := range shared {
			set[v] = true
		}
		for v := range column {
			set[v] = true
		}
		values := make([]string, 0, len(set))
		for v := range set {
			values = append(values, v)
		}
		sort.Strings(values)

		var next [][]string
		for _, prefix := range requests {
			for _, v := range values {
				next = append(next, append(append([]string{}, prefix...), v))
			}
		}
		requests = conformanceSample(next, 100000)
	}
	return requests
}

func conformanceSample(tests [][]string, limit int) [][]string {
	if len(tests) <= limit {
		return tests
	}
	stride := float64(len(tests)) / float64(limit)
	sampled := make([][]string, 0, limit)
	for i := 0; i < limit; i++ {
		sampled = append(sampled, tests[int(float64(i)*stride)])
	}
	return sampled
}

func conformanceSelect(tests []conformanceTest) []conformanceTest {
	var allowed, denied [][]string
	byKey := map[string]conformanceTest{}
	for _, test := range tests {
		key := strings.Join(test.Request, "\x00")
		byKey[key] = test
		if test.Allow {
			allowed = append(allowed, test.Request)
		} else {
			denied = append(denied, test.Request)
		}
	}
	allowed = conformanceSample(allowed, maxConformanceTests/2)
	denied = conformanceSample(denied, maxConformanceTests-len(allowed))

	var selected []conformanceTest
	for _, request := range append(allowed, denied...) {
		selected = append(selected, byKey[strings.Join(request, "\x00")])
	}
	sort.Slice(selected, func(i, j int) bool {
		return strings.Join(selected[i].Request, "\x00") < strings.Join(selected[j].Request, "\x00")
	})
	return selected
}

func writeConformanceFile(t *testing.T, path string, data []byte) {
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func generateConformanceCases(t *testing.T) {
	for _, dir := range []string{"conformance/data", "conformance/cases"} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	for _, source := range conformanceSources {
		modelFile, policyFile := source[0], source[1]
		for _, f := range source {
			data, err := os.ReadFile(filepath.Join("examples", f))
			if err != nil {
				t.Fatal(err)
			}
			writeConformanceFile(t, filepath.Join("conformance/data", f), bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n")))
		}

		e, err := NewEnforcer(filepath.Join("examples", modelFile), filepath.Join("examples", policyFile))
		if err != nil {
			t.Fatalf("%s %s: %v", modelFile, policyFile, err)
		}

		var tests []conformanceTest
		for _, request := range conformanceCandidates(e, conformanceExtraValues[modelFile]) {
			rvals := make([]interface{}, len(request))
			for i, v := range request {
				rvals[i] = v
			}
			allow, explain, enforceErr := e.EnforceEx(rvals...)
			if enforceErr != nil {
				continue
			}
			if explain == nil {
				explain = []string{}
			}
			tests = append(tests, conformanceTest{Request: request, Allow: allow, Explain: explain})
		}
		c := conformanceCase{Model: modelFile, Policy: policyFile, Tests: conformanceSelect(tests)}

		data, err := json.MarshalIndent(c, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		writeConformanceFile(t, filepath.Join("conformance/cases", conformanceCaseName(modelFile, policyFile)+".json"), append(data, '\n'))
	}
}

func TestConformance(t *testing.T) {
	if *updateConformance {
		generateConformanceCases(t)
	}

	files, err := filepath.Glob("conformance/cases/*.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		var c conformanceCase
		if err = json.Unmarshal(data, &c); err != nil {
			t.Fatalf("%s: %v", file, err)
		}

		t.Run(strings.TrimSuffix(filepath.Base(file), ".json"), func(t *testing.T) {
			e, err := NewEnforcer(filepath.Join("conformance/data", c.Model), filepath.Join("conformance/data", c.Policy))
			if err != nil {
				t.Fatal(err)
			}
			for _, test := range c.Tests {
				rvals := make([]interface{}, len(test.Request))
				for i, v := range test.Request {
					rvals[i] = v
				}
				allow, explain, err := e.EnforceEx(rvals...)
				if err != nil {
					t.Errorf("%v: %v", test.Request, err)
					continue
				}
				if explain == nil {
					explain = []string{}
				}
				if allow != test.Allow || !reflect.DeepEqual(explain, test.Explain) {
					t.Errorf("%v: got (%v, %v), want (%v, %v)", test.Request, allow, explain, test.Allow, test.Explain)
				}
			}
		})
	}
}
