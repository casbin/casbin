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
	"fmt"
	"net"
	"path"
	"regexp"
	"strings"
	"sync"

	"github.com/Knetic/govaluate"
	"github.com/casbin/casbin/v2/rbac"
)

var (
	keyMatch4Re *regexp.Regexp = regexp.MustCompile(`{([^/]+)}`)
)

// validate the variadic parameter size and type as string
func validateVariadicArgs(expectedLen int, args ...interface{}) error {
	if len(args) != expectedLen {
		return fmt.Errorf("Expected %d arguments, but got %d", expectedLen, len(args))
	}

	for _, p := range args {
		_, ok := p.(string)
		if !ok {
			return errors.New("Argument must be a string")
		}
	}

	return nil
}

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
	if err := validateVariadicArgs(2, args...); err != nil {
		return false, fmt.Errorf("%s: %s", "keyMatch", err)
	}

	name1 := args[0].(string)
	name2 := args[1].(string)

	return bool(KeyMatch(name1, name2)), nil
}

// KeyGet returns the matched part
// For example, "/foo/bar/foo" matches "/foo/*"
// "bar/foo" will been returned
func KeyGet(key1, key2 string) string {
	i := strings.Index(key2, "*")
	if i == -1 {
		return ""
	}
	if len(key1) > i {
		if key1[:i] == key2[:i] {
			return key1[i:]
		}
	}
	return ""
}

// KeyGetFunc is the wrapper for KeyGet
func KeyGetFunc(args ...interface{}) (interface{}, error) {
	if err := validateVariadicArgs(2, args...); err != nil {
		return false, fmt.Errorf("%s: %s", "keyGet", err)
	}

	name1 := args[0].(string)
	name2 := args[1].(string)

	return KeyGet(name1, name2), nil
}

// KeyMatch2 determines whether key1 matches the pattern of key2 (similar to RESTful path), key2 can contain a *.
// For example, "/foo/bar" matches "/foo/*", "/resource1" matches "/:resource"
func KeyMatch2(key1 string, key2 string) bool {
	key2 = strings.Replace(key2, "/*", "/.*", -1)

	re := regexp.MustCompile(`:[^/]+`)
	key2 = re.ReplaceAllString(key2, "$1[^/]+$2")

	return RegexMatch(key1, "^"+key2+"$")
}

// KeyMatch2Func is the wrapper for KeyMatch2.
func KeyMatch2Func(args ...interface{}) (interface{}, error) {
	if err := validateVariadicArgs(2, args...); err != nil {
		return false, fmt.Errorf("%s: %s", "keyMatch2", err)
	}

	name1 := args[0].(string)
	name2 := args[1].(string)

	return bool(KeyMatch2(name1, name2)), nil
}

// KeyGet2 returns value matched pattern
// For example, "/resource1" matches "/:resource"
// if the pathVar == "resource", then "resource1" will be returned
func KeyGet2(key1, key2 string, pathVar string) string {
	key2 = strings.Replace(key2, "/*", "/.*", -1)

	re := regexp.MustCompile(`:[^/]+`)
	keys := re.FindAllString(key2, -1)
	key2 = re.ReplaceAllString(key2, "$1([^/]+)$2")
	key2 = "^" + key2 + "$"
	re2 := regexp.MustCompile(key2)
	values := re2.FindAllStringSubmatch(key1, -1)
	if len(values) == 0 {
		return ""
	}
	for i, key := range keys {
		if pathVar == key[1:] {
			return values[0][i+1]
		}
	}
	return ""
}

// KeyGet2Func is the wrapper for KeyGet2
func KeyGet2Func(args ...interface{}) (interface{}, error) {
	if err := validateVariadicArgs(3, args...); err != nil {
		return false, fmt.Errorf("%s: %s", "keyGet2", err)
	}

	name1 := args[0].(string)
	name2 := args[1].(string)
	key := args[2].(string)

	return KeyGet2(name1, name2, key), nil
}

// KeyMatch3 determines whether key1 matches the pattern of key2 (similar to RESTful path), key2 can contain a *.
// For example, "/foo/bar" matches "/foo/*", "/resource1" matches "/{resource}"
func KeyMatch3(key1 string, key2 string) bool {
	key2 = strings.Replace(key2, "/*", "/.*", -1)

	re := regexp.MustCompile(`\{[^/]+\}`)
	key2 = re.ReplaceAllString(key2, "$1[^/]+$2")

	return RegexMatch(key1, "^"+key2+"$")
}

// KeyMatch3Func is the wrapper for KeyMatch3.
func KeyMatch3Func(args ...interface{}) (interface{}, error) {
	if err := validateVariadicArgs(2, args...); err != nil {
		return false, fmt.Errorf("%s: %s", "keyMatch3", err)
	}

	name1 := args[0].(string)
	name2 := args[1].(string)

	return bool(KeyMatch3(name1, name2)), nil
}

// KeyGet3 returns value matched pattern
// For example, "project/proj_project1_admin/" matches "project/proj_{project}_admin/"
// if the pathVar == "project", then "project1" will be returned
func KeyGet3(key1, key2 string, pathVar string) string {
	key2 = strings.Replace(key2, "/*", "/.*", -1)

	re := regexp.MustCompile(`\{[^/]+?\}`) // non-greedy match of `{...}` to support multiple {} in `/.../`
	keys := re.FindAllString(key2, -1)
	key2 = re.ReplaceAllString(key2, "$1([^/]+?)$2")
	key2 = "^" + key2 + "$"
	re2 := regexp.MustCompile(key2)
	values := re2.FindAllStringSubmatch(key1, -1)
	if len(values) == 0 {
		return ""
	}
	for i, key := range keys {
		if pathVar == key[1:len(key)-1] {
			return values[0][i+1]
		}
	}
	return ""
}

// KeyGet3Func is the wrapper for KeyGet3
func KeyGet3Func(args ...interface{}) (interface{}, error) {
	if err := validateVariadicArgs(3, args...); err != nil {
		return false, fmt.Errorf("%s: %s", "keyGet3", err)
	}

	name1 := args[0].(string)
	name2 := args[1].(string)
	key := args[2].(string)

	return KeyGet3(name1, name2, key), nil
}

// KeyMatch4 determines whether key1 matches the pattern of key2 (similar to RESTful path), key2 can contain a *.
// Besides what KeyMatch3 does, KeyMatch4 can also match repeated patterns:
// "/parent/123/child/123" matches "/parent/{id}/child/{id}"
// "/parent/123/child/456" does not match "/parent/{id}/child/{id}"
// But KeyMatch3 will match both.
func KeyMatch4(key1 string, key2 string) bool {
	key2 = strings.Replace(key2, "/*", "/.*", -1)

	tokens := []string{}

	re := keyMatch4Re
	key2 = re.ReplaceAllStringFunc(key2, func(s string) string {
		tokens = append(tokens, s[1:len(s)-1])
		return "([^/]+)"
	})

	re = regexp.MustCompile("^" + key2 + "$")
	matches := re.FindStringSubmatch(key1)
	if matches == nil {
		return false
	}
	matches = matches[1:]

	if len(tokens) != len(matches) {
		panic(errors.New("KeyMatch4: number of tokens is not equal to number of values"))
	}

	values := map[string]string{}

	for key, token := range tokens {
		if _, ok := values[token]; !ok {
			values[token] = matches[key]
		}
		if values[token] != matches[key] {
			return false
		}
	}

	return true
}

// KeyMatch4Func is the wrapper for KeyMatch4.
func KeyMatch4Func(args ...interface{}) (interface{}, error) {
	if err := validateVariadicArgs(2, args...); err != nil {
		return false, fmt.Errorf("%s: %s", "keyMatch4", err)
	}

	name1 := args[0].(string)
	name2 := args[1].(string)

	return bool(KeyMatch4(name1, name2)), nil
}

// KeyMatch5 determines whether key1 matches the pattern of key2 (similar to RESTful path), key2 can contain a *
// For example, 
// - "/foo/bar?status=1&type=2" matches "/foo/bar"
// - "/parent/child1" and "/parent/child1" matches "/parent/*"
// - "/parent/child1?status=1" matches "/parent/*"
func KeyMatch5(key1 string, key2 string) bool {
	i := strings.Index(key1, "?")

	if i != -1 {
		key1 = key1[:i]
	}

	key2 = strings.Replace(key2, "/*", "/.*", -1)

	re := regexp.MustCompile(`\{[^/]+\}`)
	key2 = re.ReplaceAllString(key2, "$1[^/]+$2")

	return RegexMatch(key1, "^"+key2+"$")
}

// KeyMatch5Func is the wrapper for KeyMatch5.
func KeyMatch5Func(args ...interface{}) (interface{}, error) {
	if err := validateVariadicArgs(2, args...); err != nil {
		return false, fmt.Errorf("%s: %s", "keyMatch5", err)
	}

	name1 := args[0].(string)
	name2 := args[1].(string)

	return bool(KeyMatch5(name1, name2)), nil
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
	if err := validateVariadicArgs(2, args...); err != nil {
		return false, fmt.Errorf("%s: %s", "regexMatch", err)
	}

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
	if err := validateVariadicArgs(2, args...); err != nil {
		return false, fmt.Errorf("%s: %s", "ipMatch", err)
	}

	ip1 := args[0].(string)
	ip2 := args[1].(string)

	return bool(IPMatch(ip1, ip2)), nil
}

// GlobMatch determines whether key1 matches the pattern of key2 using glob pattern
func GlobMatch(key1 string, key2 string) (bool, error) {
	return path.Match(key2, key1)
}

// GlobMatchFunc is the wrapper for GlobMatch.
func GlobMatchFunc(args ...interface{}) (interface{}, error) {
	if err := validateVariadicArgs(2, args...); err != nil {
		return false, fmt.Errorf("%s: %s", "globMatch", err)
	}

	name1 := args[0].(string)
	name2 := args[1].(string)

	return GlobMatch(name1, name2)
}

// GenerateGFunction is the factory method of the g(_, _[, _]) function.
func GenerateGFunction(rm rbac.RoleManager) govaluate.ExpressionFunction {
	memorized := sync.Map{}
	return func(args ...interface{}) (interface{}, error) {
		// Like all our other govaluate functions, all args are strings.

		// Allocate and generate a cache key from the arguments...
		total := len(args)
		for _, a := range args {
			aStr := a.(string)
			total += len(aStr)
		}
		builder := strings.Builder{}
		builder.Grow(total)
		for _, arg := range args {
			builder.WriteByte(0)
			builder.WriteString(arg.(string))
		}
		key := builder.String()

		// ...and see if we've already calculated this.
		v, found := memorized.Load(key)
		if found {
			return v, nil
		}

		// If not, do the calculation.
		// There are guaranteed to be exactly 2 or 3 arguments.
		name1, name2 := args[0].(string), args[1].(string)
		if rm == nil {
			v = name1 == name2
		} else if len(args) == 2 {
			v, _ = rm.HasLink(name1, name2)
		} else {
			domain := args[2].(string)
			v, _ = rm.HasLink(name1, name2, domain)
		}

		memorized.Store(key, v)
		return v, nil
	}
}
