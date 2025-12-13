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
	"strings"
)

var (
	// Regex to detect block-style matcher (starts with {)
	blockMatcherRegex = regexp.MustCompile(`^\s*\{`)
)

const (
	// maxSubstitutionPasses defines the maximum number of variable substitution passes
	// to prevent infinite loops in case of circular references
	maxSubstitutionPasses = 10
)

// TransformBlockMatcher transforms a block-style matcher to a single-line expression
// that can be evaluated by govaluate.
//
// Example transformation:
// Input:
//   {
//     let role_match = g(r.sub, p.sub)
//     let obj_match = r.obj == p.obj
//     return role_match && obj_match
//   }
// Output:
//   g(r.sub, p.sub) && r.obj == p.obj
func TransformBlockMatcher(matcher string) string {
	matcher = strings.TrimSpace(matcher)
	
	// Check if this is a block-style matcher
	if !blockMatcherRegex.MatchString(matcher) {
		return matcher
	}

	// Remove outer braces
	matcher = strings.TrimPrefix(matcher, "{")
	matcher = strings.TrimSuffix(strings.TrimSpace(matcher), "}")
	matcher = strings.TrimSpace(matcher)

	// Parse the block into statements
	statements := parseStatements(matcher)
	
	// Build a map of variable substitutions from let statements
	varMap := make(map[string]string)
	var ifStatements []ifStatement
	var finalReturn string
	
	for _, stmt := range statements {
		if stmt.stmtType == stmtTypeLet {
			varMap[stmt.varName] = stmt.expression
		} else if stmt.stmtType == stmtTypeIf {
			ifStatements = append(ifStatements, ifStatement{
				condition:   stmt.condition,
				returnValue: stmt.expression,
			})
		} else if stmt.stmtType == stmtTypeReturn {
			finalReturn = stmt.expression
		}
	}

	// Substitute variables in all expressions
	substituteVars := func(expr string) string {
		// Perform multiple passes to handle nested variable references
		for i := 0; i < maxSubstitutionPasses; i++ {
			changed := false
			for varName, varExpr := range varMap {
				// Use word boundaries to avoid partial matches
				pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(varName) + `\b`)
				newExpr := pattern.ReplaceAllString(expr, "("+varExpr+")")
				if newExpr != expr {
					changed = true
					expr = newExpr
				}
			}
			if !changed {
				break
			}
		}
		return expr
	}

	// Substitute variables in if conditions and return values
	for i := range ifStatements {
		ifStatements[i].condition = substituteVars(ifStatements[i].condition)
		ifStatements[i].returnValue = substituteVars(ifStatements[i].returnValue)
	}
	finalReturn = substituteVars(finalReturn)

	// Build the final expression
	// Handle early returns by converting them to conditional logic
	result := finalReturn
	for i := len(ifStatements) - 1; i >= 0; i-- {
		condition := ifStatements[i].condition
		returnValue := ifStatements[i].returnValue
		// Transform: if condition { return returnValue } else { ... rest ... }
		// to: (condition && returnValue) || (!condition && rest)
		result = "((" + condition + ") && (" + returnValue + ")) || (!(" + condition + ") && (" + result + "))"
	}

	return result
}

type statementType int

const (
	stmtTypeLet statementType = iota
	stmtTypeIf
	stmtTypeReturn
)

type statement struct {
	stmtType   statementType
	varName    string
	expression string
	condition  string
}

type ifStatement struct {
	condition   string
	returnValue string
}

func parseStatements(block string) []statement {
	var statements []statement
	
	// Split by keywords: let, if, return
	// We need to be careful about parsing
	i := 0
	for i < len(block) {
		// Skip whitespace
		for i < len(block) && (block[i] == ' ' || block[i] == '\t' || block[i] == '\n' || block[i] == '\r') {
			i++
		}
		if i >= len(block) {
			break
		}
		
		// Check for keywords
		if strings.HasPrefix(block[i:], "let ") {
			// Parse let statement
			i += 4 // skip "let "
			// Find variable name
			varStart := i
			for i < len(block) && (isLetterOrDigit(block[i]) || block[i] == '_') {
				i++
			}
			varName := block[varStart:i]
			
			// Skip whitespace and '='
			for i < len(block) && (block[i] == ' ' || block[i] == '\t' || block[i] == '=') {
				i++
			}
			
			// Find expression (until next keyword or end)
			exprStart := i
			depth := 0
			for i < len(block) {
				if block[i] == '(' || block[i] == '[' || block[i] == '{' {
					depth++
				} else if block[i] == ')' || block[i] == ']' || block[i] == '}' {
					depth--
				}
				
				if depth == 0 {
					// Check if we're at the start of a keyword
					remaining := block[i:]
					if strings.HasPrefix(remaining, "let ") || 
					   strings.HasPrefix(remaining, "if ") || 
					   strings.HasPrefix(remaining, "return ") {
						break
					}
				}
				i++
			}
			expression := strings.TrimSpace(block[exprStart:i])
			
			statements = append(statements, statement{
				stmtType:   stmtTypeLet,
				varName:    varName,
				expression: expression,
			})
			
		} else if strings.HasPrefix(block[i:], "if ") {
			// Parse if statement with return
			i += 3 // skip "if "
			
			// Find condition (until '{')
			condStart := i
			for i < len(block) && block[i] != '{' {
				i++
			}
			condition := strings.TrimSpace(block[condStart:i])
			
			// Skip '{'
			if i < len(block) && block[i] == '{' {
				i++
			}
			
			// Skip whitespace and "return"
			for i < len(block) && (block[i] == ' ' || block[i] == '\t' || block[i] == '\n' || block[i] == '\r') {
				i++
			}
			if strings.HasPrefix(block[i:], "return ") {
				i += 7 // skip "return "
			}
			
			// Find return value (until '}')
			valueStart := i
			for i < len(block) && block[i] != '}' {
				i++
			}
			returnValue := strings.TrimSpace(block[valueStart:i])
			
			// Skip '}'
			if i < len(block) && block[i] == '}' {
				i++
			}
			
			statements = append(statements, statement{
				stmtType:   stmtTypeIf,
				condition:  condition,
				expression: returnValue,
			})
			
		} else if strings.HasPrefix(block[i:], "return ") {
			// Parse return statement
			i += 7 // skip "return "
			
			// Find expression (until end)
			exprStart := i
			i = len(block)
			expression := strings.TrimSpace(block[exprStart:i])
			
			statements = append(statements, statement{
				stmtType:   stmtTypeReturn,
				expression: expression,
			})
			
		} else {
			// Skip whitespace - this is expected
			if i < len(block) && (block[i] == ' ' || block[i] == '\t' || block[i] == '\n' || block[i] == '\r') {
				i++
			} else if i < len(block) {
				// Unknown token - this could indicate a syntax error
				// Skip the character and continue parsing
				i++
			}
		}
	}
	
	return statements
}

func isLetterOrDigit(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}
