// Copyright 2017 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"errors"
	"net"
	"regexp"
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/casbin/casbin/v2/rbac"
)

// KeyMatch determines whether key1 matches the pattern of key2 (similar to RESTful path), key2 can contain a *.
// For example, "/foo/bar" matches "/foo/*"
func KeyMatch(key1 string, key2 string) bool {
	i := strings.Index(key2, "*")
	if i == -1 {
		return key1 == key2
	}

	if len(key1) > i {
		return key1[:i] == key2[:i]
	}
	return key1 == key2[:i]
}

// KeyMatchFunc is the wrapper for KeyMatch.
func KeyMatchFunc(args ...interface{}) (interface{}, error) {
	name1 := args[0].(string)
	name2 := args[1].(string)

	return bool(KeyMatch(name1, name2)), nil
}

// KeyMatch2 determines whether key1 matches the pattern of key2 (similar to RESTful path), key2 can contain a *.
// For example, "/foo/bar" matches "/foo/*", "/resource1" matches "/:resource"
func KeyMatch2(key1 string, key2 string) bool {
	key2 = strings.Replace(key2, "/*", "/.*", -1)

	re := regexp.MustCompile(`(.*):[^/]+(.*)`)
	for {
		if !strings.Contains(key2, "/:") {
			break
		}

		key2 = re.ReplaceAllString(key2, "$1[^/]+$2")
	}

	return RegexMatch(key1, "^"+key2+"$")
}

// KeyMatch2Func is the wrapper for KeyMatch2.
func KeyMatch2Func(args ...interface{}) (interface{}, error) {
	name1 := args[0].(string)
	name2 := args[1].(string)

	return bool(KeyMatch2(name1, name2)), nil
}

// KeyMatch3 determines whether key1 matches the pattern of key2 (similar to RESTful path), key2 can contain a *.
// For example, "/foo/bar" matches "/foo/*", "/resource1" matches "/{resource}"
func KeyMatch3(key1 string, key2 string) bool {
	key2 = strings.Replace(key2, "/*", "/.*", -1)

	re := regexp.MustCompile(`(.*)\{[^/]+\}(.*)`)
	for {
		if !strings.Contains(key2, "/{") {
			break
		}

		key2 = re.ReplaceAllString(key2, "$1[^/]+$2")
	}

	return RegexMatch(key1, "^"+key2+"$")
}

// KeyMatch3Func is the wrapper for KeyMatch3.
func KeyMatch3Func(args ...interface{}) (interface{}, error) {
	name1 := args[0].(string)
	name2 := args[1].(string)

	return bool(KeyMatch3(name1, name2)), nil
}

// KeyMatch4 determines whether key1 matches the pattern of key2 (similar to RESTful path), key2 can contain a *.
// Besides what KeyMatch3 does, KeyMatch4 can also match repeated patterns:
// "/parent/123/child/123" matches "/parent/{id}/child/{id}"
// "/parent/123/child/456" does not match "/parent/{id}/child/{id}"
// But KeyMatch3 will match both.
func KeyMatch4(key1 string, key2 string) bool {
	key2 = strings.Replace(key2, "/*", "/.*", -1)

	tokens := []string{}
	j := -1
	for i, c := range key2 {
		if c == '{' {
			j = i
		} else if c == '}' {
			tokens = append(tokens, key2[j:i+1])
		}
	}

	re := regexp.MustCompile(`(.*)\{[^/]+\}(.*)`)
	for {
		if !strings.Contains(key2, "/{") {
			break
		}

		key2 = re.ReplaceAllString(key2, "$1([^/]+)$2")
	}

	re = regexp.MustCompile("^" + key2 + "$")
	values := re.FindStringSubmatch(key1)
	if values == nil {
		return false
	}
	values = values[1:]

	if len(tokens) != len(values) {
		panic(errors.New("KeyMatch4: number of tokens is not equal to number of values"))
	}

	m := map[string][]string{}
	for i := 0; i < len(tokens); i++ {
		if _, ok := m[tokens[i]]; !ok {
			m[tokens[i]] = []string{}
		}

		m[tokens[i]] = append(m[tokens[i]], values[i])
	}

	for _, values := range m {
		if len(values) > 1 {
			for i := 1; i < len(values); i++ {
				if values[i] != values[0] {
					return false
				}
			}
		}
	}

	return true
}

// KeyMatch4Func is the wrapper for KeyMatch4.
func KeyMatch4Func(args ...interface{}) (interface{}, error) {
	name1 := args[0].(string)
	name2 := args[1].(string)

	return bool(KeyMatch4(name1, name2)), nil
}

// RegexMatch determines whether key1 matches the pattern of key2 in regular expression.
func RegexMatch(key1 string, key2 string) bool {
	res, err := regexp.MatchString(key2, key1)
	if err != nil {
		panic(err)
	}
	return res
}

// RegexMatchFunc is the wrapper for RegexMatch.
func RegexMatchFunc(args ...interface{}) (interface{}, error) {
	name1 := args[0].(string)
	name2 := args[1].(string)

	return bool(RegexMatch(name1, name2)), nil
}

// IPMatch determines whether IP address ip1 matches the pattern of IP address ip2, ip2 can be an IP address or a CIDR pattern.
// For example, "192.168.2.123" matches "192.168.2.0/24"
func IPMatch(ip1 string, ip2 string) bool {
	objIP1 := net.ParseIP(ip1)
	if objIP1 == nil {
		panic("invalid argument: ip1 in IPMatch() function is not an IP address.")
	}

	_, cidr, err := net.ParseCIDR(ip2)
	if err != nil {
		objIP2 := net.ParseIP(ip2)
		if objIP2 == nil {
			panic("invalid argument: ip2 in IPMatch() function is neither an IP address nor a CIDR.")
		}

		return objIP1.Equal(objIP2)
	}

	return cidr.Contains(objIP1)
}

// IPMatchFunc is the wrapper for IPMatch.
func IPMatchFunc(args ...interface{}) (interface{}, error) {
	ip1 := args[0].(string)
	ip2 := args[1].(string)

	return bool(IPMatch(ip1, ip2)), nil
}

// GenerateGFunction is the factory method of the g(_, _) function.
func GenerateGFunction(rm rbac.RoleManager) govaluate.ExpressionFunction {
	return func(args ...interface{}) (interface{}, error) {
		name1 := args[0].(string)
		name2 := args[1].(string)

		if rm == nil {
			return name1 == name2, nil
		} else if len(args) == 2 {
			res, _ := rm.HasLink(name1, name2)
			return res, nil
		} else {
			domain := args[2].(string)
			res, _ := rm.HasLink(name1, name2, domain)
			return res, nil
		}
	}
}
