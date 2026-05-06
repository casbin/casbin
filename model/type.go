// Copyright 2026 The casbin Authors. All Rights Reserved.
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

package model

import (
	"fmt"
	"strings"

	"github.com/casbin/casbin/v3/constant"
	Err "github.com/casbin/casbin/v3/errors"
)

const (
	typeDefinitionSection = "t"
	userTypeKey           = "user"
	roleTypeKey           = "role"
)

type entityType string

const (
	entityTypeUnknown entityType = ""
	entityTypeUser    entityType = "user"
	entityTypeRole    entityType = "role"
)

type typeDefinition struct {
	userPrefix string
	rolePrefix string
}

func (model Model) getTypeDefinition() (*typeDefinition, bool, error) {
	section := model[typeDefinitionSection]
	if section == nil || len(section) == 0 {
		return nil, false, nil
	}

	userAssertion, hasUser := section[userTypeKey]
	roleAssertion, hasRole := section[roleTypeKey]
	if !hasUser || !hasRole {
		return nil, false, fmt.Errorf("%w: type_definition must define both user and role", Err.ErrInvalidTypeDefinition)
	}

	userPrefix := strings.TrimSpace(userAssertion.Value)
	rolePrefix := strings.TrimSpace(roleAssertion.Value)
	if userPrefix == "" || rolePrefix == "" {
		return nil, false, fmt.Errorf("%w: user and role prefixes cannot be empty", Err.ErrInvalidTypeDefinition)
	}
	if userPrefix == rolePrefix {
		return nil, false, fmt.Errorf("%w: user and role prefixes must be different", Err.ErrInvalidTypeDefinition)
	}
	if strings.HasPrefix(userPrefix, rolePrefix) || strings.HasPrefix(rolePrefix, userPrefix) {
		return nil, false, fmt.Errorf("%w: user and role prefixes must not overlap", Err.ErrInvalidTypeDefinition)
	}

	return &typeDefinition{userPrefix: userPrefix, rolePrefix: rolePrefix}, true, nil
}

func (model Model) ValidateTypeDefinitions() error {
	_, _, err := model.getTypeDefinition()
	return err
}

func (model Model) GetEntityType(name string) (string, bool, error) {
	def, enabled, err := model.getTypeDefinition()
	if err != nil || !enabled {
		return "", enabled, err
	}

	switch {
	case strings.HasPrefix(name, def.userPrefix):
		return string(entityTypeUser), true, nil
	case strings.HasPrefix(name, def.rolePrefix):
		return string(entityTypeRole), true, nil
	default:
		return "", true, nil
	}
}

func (model Model) ValidatePolicyTypes(sec string, ptype string, rule []string) error {
	def, enabled, err := model.getTypeDefinition()
	if err != nil || !enabled {
		return err
	}

	switch sec {
	case "p":
		index, err := model.GetFieldIndex(ptype, constant.SubjectIndex)
		if err != nil || index >= len(rule) {
			return nil
		}
		return validateEntityType(rule[index], ptype+".sub", def, entityTypeUser, entityTypeRole)
	case "g":
		if ptype != "g" || len(rule) < 2 {
			return nil
		}
		if err := validateEntityType(rule[0], ptype+"[0]", def, entityTypeUser, entityTypeRole); err != nil {
			return err
		}
		return validateEntityType(rule[1], ptype+"[1]", def, entityTypeRole)
	default:
		return nil
	}
}

func validateEntityType(name string, field string, def *typeDefinition, allowed ...entityType) error {
	actual := getEntityType(name, def)
	if actual == entityTypeUnknown {
		return fmt.Errorf("type mismatch for %s: %q does not match any configured user/role prefix", field, name)
	}

	for _, allowedType := range allowed {
		if actual == allowedType {
			return nil
		}
	}

	expected := make([]string, 0, len(allowed))
	for _, allowedType := range allowed {
		expected = append(expected, string(allowedType))
	}

	return fmt.Errorf("type mismatch for %s: %q is %s, expected %s", field, name, actual, strings.Join(expected, " or "))
}

func getEntityType(name string, def *typeDefinition) entityType {
	switch {
	case strings.HasPrefix(name, def.userPrefix):
		return entityTypeUser
	case strings.HasPrefix(name, def.rolePrefix):
		return entityTypeRole
	default:
		return entityTypeUnknown
	}
}
