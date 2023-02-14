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
	"regexp"
	"sort"
	"strings"
	"sync"
)

var evalReg = regexp.MustCompile(`\beval\((?P<rule>[^)]*)\)`)

var escapeAssertionRegex = regexp.MustCompile(`\b((r|p)[0-9]*)\.`)

// EscapeAssertion escapes the dots in the assertion, because the expression evaluation doesn't support such variable names.
func EscapeAssertion(s string) string {
	s = escapeAssertionRegex.ReplaceAllStringFunc(s, func(m string) string {
		return strings.Replace(m, ".", "_", 1)
	})
	return s
}

// RemoveComments removes the comments starting with # in the text.
func RemoveComments(s string) string {
	pos := strings.Index(s, "#")
	if pos == -1 {
		return s
	}
	return strings.TrimSpace(s[0:pos])
}

// ArrayEquals determines whether two string arrays are identical.
func ArrayEquals(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// Array2DEquals determines whether two 2-dimensional string arrays are identical.
func Array2DEquals(a [][]string, b [][]string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if !ArrayEquals(v, b[i]) {
			return false
		}
	}
	return true
}

// SortArray2D  Sorts the two-dimensional string array
func SortArray2D(arr [][]string) {
	if len(arr) != 0 {
		sort.Slice(arr, func(i, j int) bool {
			elementLen := len(arr[0])
			for k := 0; k < elementLen; k++ {
				if arr[i][k] < arr[j][k] {
					return true
				} else if arr[i][k] > arr[j][k] {
					return false
				}
			}
			return true
		})
	}
}

// SortedArray2DEquals determines whether two 2-dimensional string arrays are identical.
func SortedArray2DEquals(a [][]string, b [][]string) bool {
	if len(a) != len(b) {
		return false
	}
	copyA := make([][]string, len(a))
	copy(copyA, a)
	copyB := make([][]string, len(b))
	copy(copyB, b)

	SortArray2D(copyA)
	SortArray2D(copyB)

	for i, v := range copyA {
		if !ArrayEquals(v, copyB[i]) {
			return false
		}
	}
	return true
}

// ArrayRemoveDuplicates removes any duplicated elements in a string array.
func ArrayRemoveDuplicates(s *[]string) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *s {
		if !found[x] {
			found[x] = true
			(*s)[j] = (*s)[i]
			j++
		}
	}
	*s = (*s)[:j]
}

// ArrayToString gets a printable string for a string array.
func ArrayToString(s []string) string {
	return strings.Join(s, ", ")
}

// ParamsToString gets a printable string for variable number of parameters.
func ParamsToString(s ...string) string {
	return strings.Join(s, ", ")
}

// SetEquals determines whether two string sets are identical.
func SetEquals(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	sort.Strings(a)
	sort.Strings(b)

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// SetEquals determines whether two string sets are identical.
func SetEqualsInt(a []int, b []int) bool {
	if len(a) != len(b) {
		return false
	}

	sort.Ints(a)
	sort.Ints(b)

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// SetEquals determines whether two string sets are identical.
func Set2DEquals(a [][]string, b [][]string) bool {
	if len(a) != len(b) {
		return false
	}

	var aa []string
	for _, v := range a {
		sort.Strings(v)
		aa = append(aa, strings.Join(v, ", "))
	}
	var bb []string
	for _, v := range b {
		sort.Strings(v)
		bb = append(bb, strings.Join(v, ", "))
	}

	return SetEquals(aa, bb)
}

// JoinSlice joins a string and a slice into a new slice.
func JoinSlice(a string, b ...string) []string {
	res := make([]string, 0, len(b)+1)

	res = append(res, a)
	res = append(res, b...)

	return res
}

// JoinSliceAny joins a string and a slice into a new interface{} slice.
func JoinSliceAny(a string, b ...string) []interface{} {
	res := make([]interface{}, 0, len(b)+1)

	res = append(res, a)
	for _, s := range b {
		res = append(res, s)
	}

	return res
}

// SetSubtract returns the elements in `a` that aren't in `b`.
func SetSubtract(a []string, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

// HasEval determine whether matcher contains function eval
func HasEval(s string) bool {
	return evalReg.MatchString(s)
}

// ReplaceEval replace function eval with the value of its parameters
func ReplaceEval(s string, rule string) string {
	return evalReg.ReplaceAllString(s, "("+rule+")")
}

// ReplaceEvalWithMap replace function eval with the value of its parameters via given sets.
func ReplaceEvalWithMap(src string, sets map[string]string) string {
	return evalReg.ReplaceAllStringFunc(src, func(s string) string {
		subs := evalReg.FindStringSubmatch(s)
		if subs == nil {
			return s
		}
		key := subs[1]
		value, found := sets[key]
		if !found {
			return s
		}
		return evalReg.ReplaceAllString(s, value)
	})
}

// GetEvalValue returns the parameters of function eval
func GetEvalValue(s string) []string {
	subMatch := evalReg.FindAllStringSubmatch(s, -1)
	var rules []string
	for _, rule := range subMatch {
		rules = append(rules, rule[1])
	}
	return rules
}

func RemoveDuplicateElement(s []string) []string {
	result := make([]string, 0, len(s))
	temp := map[string]struct{}{}
	for _, item := range s {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

type node struct {
	key   interface{}
	value interface{}
	prev  *node
	next  *node
}

type LRUCache struct {
	capacity int
	m        map[interface{}]*node
	head     *node
	tail     *node
}

func NewLRUCache(capacity int) *LRUCache {
	cache := &LRUCache{}
	cache.capacity = capacity
	cache.m = map[interface{}]*node{}

	head := &node{}
	tail := &node{}

	head.next = tail
	tail.prev = head

	cache.head = head
	cache.tail = tail

	return cache
}

func (cache *LRUCache) remove(n *node, listOnly bool) {
	if !listOnly {
		delete(cache.m, n.key)
	}
	n.prev.next = n.next
	n.next.prev = n.prev
}

func (cache *LRUCache) add(n *node, listOnly bool) {
	if !listOnly {
		cache.m[n.key] = n
	}
	headNext := cache.head.next
	cache.head.next = n
	headNext.prev = n
	n.next = headNext
	n.prev = cache.head
}

func (cache *LRUCache) moveToHead(n *node) {
	cache.remove(n, true)
	cache.add(n, true)
}

func (cache *LRUCache) Get(key interface{}) (value interface{}, ok bool) {
	n, ok := cache.m[key]
	if ok {
		cache.moveToHead(n)
		return n.value, ok
	} else {
		return nil, ok
	}
}

func (cache *LRUCache) Put(key interface{}, value interface{}) {
	n, ok := cache.m[key]
	if ok {
		cache.remove(n, false)
	} else {
		n = &node{key, value, nil, nil}
		if len(cache.m) >= cache.capacity {
			cache.remove(cache.tail.prev, false)
		}
	}
	cache.add(n, false)
}

type SyncLRUCache struct {
	rwm sync.RWMutex
	*LRUCache
}

func NewSyncLRUCache(capacity int) *SyncLRUCache {
	cache := &SyncLRUCache{}
	cache.LRUCache = NewLRUCache(capacity)
	return cache
}

func (cache *SyncLRUCache) Get(key interface{}) (value interface{}, ok bool) {
	cache.rwm.Lock()
	defer cache.rwm.Unlock()
	return cache.LRUCache.Get(key)
}

func (cache *SyncLRUCache) Put(key interface{}, value interface{}) {
	cache.rwm.Lock()
	defer cache.rwm.Unlock()
	cache.LRUCache.Put(key, value)
}
