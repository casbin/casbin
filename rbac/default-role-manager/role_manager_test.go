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

package defaultrolemanager

import (
	"log"
	"testing"

	"github.com/casbin/casbin/rbac"
	"github.com/casbin/casbin/util"
)

func testRole(t *testing.T, rm rbac.RoleManager, name1 string, name2 string, res bool) {
	t.Helper()
	myRes, _ := rm.HasLink(name1, name2)
	log.Printf("%s, %s: %t", name1, name2, myRes)

	if myRes != res {
		t.Errorf("%s < %s: %t, supposed to be %t", name1, name2, !res, res)
	}
}

func testDomainRole(t *testing.T, rm rbac.RoleManager, name1 string, name2 string, domain string, res bool) {
	t.Helper()
	myRes, _ := rm.HasLink(name1, name2, domain)
	log.Printf("%s :: %s, %s: %t", domain, name1, name2, myRes)

	if myRes != res {
		t.Errorf("%s :: %s < %s: %t, supposed to be %t", domain, name1, name2, !res, res)
	}
}

func testPrintRoles(t *testing.T, rm rbac.RoleManager, name string, res []string) {
	t.Helper()
	myRes, _ := rm.GetRoles(name)
	log.Printf("%s: %s", name, myRes)

	if !util.ArrayEquals(myRes, res) {
		t.Errorf("%s: %s, supposed to be %s", name, myRes, res)
	}
}

func TestRole(t *testing.T) {
	rm := NewRoleManager(3)
	rm.AddLink("u1", "g1")
	rm.AddLink("u2", "g1")
	rm.AddLink("u3", "g2")
	rm.AddLink("u4", "g2")
	rm.AddLink("u4", "g3")
	rm.AddLink("g1", "g3")

	// Current role inheritance tree:
	//             g3    g2
	//            /  \  /  \
	//          g1    u4    u3
	//         /  \
	//       u1    u2

	testRole(t, rm, "u1", "g1", true)
	testRole(t, rm, "u1", "g2", false)
	testRole(t, rm, "u1", "g3", true)
	testRole(t, rm, "u2", "g1", true)
	testRole(t, rm, "u2", "g2", false)
	testRole(t, rm, "u2", "g3", true)
	testRole(t, rm, "u3", "g1", false)
	testRole(t, rm, "u3", "g2", true)
	testRole(t, rm, "u3", "g3", false)
	testRole(t, rm, "u4", "g1", false)
	testRole(t, rm, "u4", "g2", true)
	testRole(t, rm, "u4", "g3", true)

	testPrintRoles(t, rm, "u1", []string{"g1"})
	testPrintRoles(t, rm, "u2", []string{"g1"})
	testPrintRoles(t, rm, "u3", []string{"g2"})
	testPrintRoles(t, rm, "u4", []string{"g2", "g3"})
	testPrintRoles(t, rm, "g1", []string{"g3"})
	testPrintRoles(t, rm, "g2", []string{})
	testPrintRoles(t, rm, "g3", []string{})

	rm.DeleteLink("g1", "g3")
	rm.DeleteLink("u4", "g2")

	// Current role inheritance tree after deleting the links:
	//             g3    g2
	//               \     \
	//          g1    u4    u3
	//         /  \
	//       u1    u2

	testRole(t, rm, "u1", "g1", true)
	testRole(t, rm, "u1", "g2", false)
	testRole(t, rm, "u1", "g3", false)
	testRole(t, rm, "u2", "g1", true)
	testRole(t, rm, "u2", "g2", false)
	testRole(t, rm, "u2", "g3", false)
	testRole(t, rm, "u3", "g1", false)
	testRole(t, rm, "u3", "g2", true)
	testRole(t, rm, "u3", "g3", false)
	testRole(t, rm, "u4", "g1", false)
	testRole(t, rm, "u4", "g2", false)
	testRole(t, rm, "u4", "g3", true)

	testPrintRoles(t, rm, "u1", []string{"g1"})
	testPrintRoles(t, rm, "u2", []string{"g1"})
	testPrintRoles(t, rm, "u3", []string{"g2"})
	testPrintRoles(t, rm, "u4", []string{"g3"})
	testPrintRoles(t, rm, "g1", []string{})
	testPrintRoles(t, rm, "g2", []string{})
	testPrintRoles(t, rm, "g3", []string{})
}

func TestDomainRole(t *testing.T) {
	rm := NewRoleManager(3)
	rm.AddLink("u1", "g1", "domain1")
	rm.AddLink("u2", "g1", "domain1")
	rm.AddLink("u3", "admin", "domain2")
	rm.AddLink("u4", "admin", "domain2")
	rm.AddLink("u4", "admin", "domain1")
	rm.AddLink("g1", "admin", "domain1")

	// Current role inheritance tree:
	//       domain1:admin    domain2:admin
	//            /       \  /       \
	//      domain1:g1     u4         u3
	//         /  \
	//       u1    u2

	testDomainRole(t, rm, "u1", "g1", "domain1", true)
	testDomainRole(t, rm, "u1", "g1", "domain2", false)
	testDomainRole(t, rm, "u1", "admin", "domain1", true)
	testDomainRole(t, rm, "u1", "admin", "domain2", false)

	testDomainRole(t, rm, "u2", "g1", "domain1", true)
	testDomainRole(t, rm, "u2", "g1", "domain2", false)
	testDomainRole(t, rm, "u2", "admin", "domain1", true)
	testDomainRole(t, rm, "u2", "admin", "domain2", false)

	testDomainRole(t, rm, "u3", "g1", "domain1", false)
	testDomainRole(t, rm, "u3", "g1", "domain2", false)
	testDomainRole(t, rm, "u3", "admin", "domain1", false)
	testDomainRole(t, rm, "u3", "admin", "domain2", true)

	testDomainRole(t, rm, "u4", "g1", "domain1", false)
	testDomainRole(t, rm, "u4", "g1", "domain2", false)
	testDomainRole(t, rm, "u4", "admin", "domain1", true)
	testDomainRole(t, rm, "u4", "admin", "domain2", true)

	rm.DeleteLink("g1", "admin", "domain1")
	rm.DeleteLink("u4", "admin", "domain2")

	// Current role inheritance tree after deleting the links:
	//       domain1:admin    domain2:admin
	//                    \          \
	//      domain1:g1     u4         u3
	//         /  \
	//       u1    u2

	testDomainRole(t, rm, "u1", "g1", "domain1", true)
	testDomainRole(t, rm, "u1", "g1", "domain2", false)
	testDomainRole(t, rm, "u1", "admin", "domain1", false)
	testDomainRole(t, rm, "u1", "admin", "domain2", false)

	testDomainRole(t, rm, "u2", "g1", "domain1", true)
	testDomainRole(t, rm, "u2", "g1", "domain2", false)
	testDomainRole(t, rm, "u2", "admin", "domain1", false)
	testDomainRole(t, rm, "u2", "admin", "domain2", false)

	testDomainRole(t, rm, "u3", "g1", "domain1", false)
	testDomainRole(t, rm, "u3", "g1", "domain2", false)
	testDomainRole(t, rm, "u3", "admin", "domain1", false)
	testDomainRole(t, rm, "u3", "admin", "domain2", true)

	testDomainRole(t, rm, "u4", "g1", "domain1", false)
	testDomainRole(t, rm, "u4", "g1", "domain2", false)
	testDomainRole(t, rm, "u4", "admin", "domain1", true)
	testDomainRole(t, rm, "u4", "admin", "domain2", false)
}

func TestClear(t *testing.T) {
	rm := NewRoleManager(3)
	rm.AddLink("u1", "g1")
	rm.AddLink("u2", "g1")
	rm.AddLink("u3", "g2")
	rm.AddLink("u4", "g2")
	rm.AddLink("u4", "g3")
	rm.AddLink("g1", "g3")

	// Current role inheritance tree:
	//             g3    g2
	//            /  \  /  \
	//          g1    u4    u3
	//         /  \
	//       u1    u2

	rm.Clear()

	// All data is cleared.
	// No role inheritance now.

	testRole(t, rm, "u1", "g1", false)
	testRole(t, rm, "u1", "g2", false)
	testRole(t, rm, "u1", "g3", false)
	testRole(t, rm, "u2", "g1", false)
	testRole(t, rm, "u2", "g2", false)
	testRole(t, rm, "u2", "g3", false)
	testRole(t, rm, "u3", "g1", false)
	testRole(t, rm, "u3", "g2", false)
	testRole(t, rm, "u3", "g3", false)
	testRole(t, rm, "u4", "g1", false)
	testRole(t, rm, "u4", "g2", false)
	testRole(t, rm, "u4", "g3", false)
}
