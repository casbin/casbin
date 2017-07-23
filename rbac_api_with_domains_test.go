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

package casbin

import (
	"log"
	"testing"

	"github.com/casbin/casbin/util"
)

func testGetRolesUnderDomain(t *testing.T, e *Enforcer, name string, domain string, res []string) {
	myRes := e.GetRolesForUserUnderDomain(name, domain)
	log.Print("Roles for", name, "under", domain, ":", myRes)

	if !util.SetEquals(res, myRes) {
		t.Error("Roles for", name, "under", domain, ":", myRes, ", supposed to be ", res)
	}
}

func TestRoleAPIWithDomains(t *testing.T) {
	e := NewEnforcer("examples/rbac_model_with_domains.conf", "examples/rbac_policy_with_domains.csv")

	testGetRolesUnderDomain(t, e, "alice", "domain1", []string{"admin"})
	testGetRolesUnderDomain(t, e, "bob", "domain1", []string{})
	testGetRolesUnderDomain(t, e, "admin", "domain1", []string{})
	testGetRolesUnderDomain(t, e, "non_exist", "domain1", []string{})

	testGetRolesUnderDomain(t, e, "alice", "domain2", []string{})
	testGetRolesUnderDomain(t, e, "bob", "domain2", []string{"admin"})
	testGetRolesUnderDomain(t, e, "admin", "domain2", []string{})
	testGetRolesUnderDomain(t, e, "non_exist", "domain2", []string{})
}