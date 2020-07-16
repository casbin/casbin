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
	"testing"

	"github.com/casbin/casbin/v3/rbac"
	"github.com/casbin/casbin/v3/util"
)

func testRole(t *testing.T, rm rbac.RoleManager, name1 string, name2 string, res bool) {
	t.Helper()
	myRes, _ := rm.HasLink(name1, name2)
	t.Logf("%s, %s: %t", name1, name2, myRes)

	if myRes != res {
		t.Errorf("%s < %s: %t, supposed to be %t", name1, name2, !res, res)
	}
}

func testDomainRole(t *testing.T, rm rbac.RoleManager, name1 string, name2 string, domain string, res bool) {
	t.Helper()
	myRes, _ := rm.HasLink(name1, name2, domain)
	t.Logf("%s :: %s, %s: %t", domain, name1, name2, myRes)

	if myRes != res {
		t.Errorf("%s :: %s < %s: %t, supposed to be %t", domain, name1, name2, !res, res)
	}
}

func testPrintRoles(t *testing.T, rm rbac.RoleManager, name string, res []string) {
	t.Helper()
	myRes, _ := rm.GetRoles(name)
	t.Logf("%s: %s", name, myRes)

	if !util.ArrayEquals(myRes, res) {
		t.Errorf("%s: %s, supposed to be %s", name, myRes, res)
	}
}

func testPrintRolesWithDomain(t *testing.T, rm rbac.RoleManager, name string, domain string, res []string) {
	t.Helper()
	myRes, _ := rm.GetRoles(name, domain)

	if !util.ArrayEquals(myRes, res) {
		t.Errorf("%s: %s, supposed to be %s", name, myRes, res)
	}
}

func TestRole(t *testing.T) {
	rm := NewRoleManager(3)
	_ = rm.AddLink("u1", "g1")
	_ = rm.AddLink("u2", "g1")
	_ = rm.AddLink("u3", "g2")
	_ = rm.AddLink("u4", "g2")
	_ = rm.AddLink("u4", "g3")
	_ = rm.AddLink("g1", "g3")

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

	_ = rm.DeleteLink("g1", "g3")
	_ = rm.DeleteLink("u4", "g2")

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
	_ = rm.AddLink("u1", "g1", "domain1")
	_ = rm.AddLink("u2", "g1", "domain1")
	_ = rm.AddLink("u3", "admin", "domain2")
	_ = rm.AddLink("u4", "admin", "domain2")
	_ = rm.AddLink("u4", "admin", "domain1")
	_ = rm.AddLink("g1", "admin", "domain1")

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

	_ = rm.DeleteLink("g1", "admin", "domain1")
	_ = rm.DeleteLink("u4", "admin", "domain2")

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
	_ = rm.AddLink("u1", "g1")
	_ = rm.AddLink("u2", "g1")
	_ = rm.AddLink("u3", "g2")
	_ = rm.AddLink("u4", "g2")
	_ = rm.AddLink("u4", "g3")
	_ = rm.AddLink("g1", "g3")

	// Current role inheritance tree:
	//             g3    g2
	//            /  \  /  \
	//          g1    u4    u3
	//         /  \
	//       u1    u2

	_ = rm.Clear()

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

func TestDomainPartternRole(t *testing.T) {
	rm := NewRoleManager(10)
	rm.(*RoleManager).AddDomainMatchingFunc("keyMatch2", util.KeyMatch2)

	_ = rm.AddLink("u1", "g1", "domain1")
	_ = rm.AddLink("u2", "g1", "domain2")
	_ = rm.AddLink("u3", "g1", "*")
	_ = rm.AddLink("u4", "g2", "domain3")
	// Current role inheritance tree after deleting the links:
	//       domain1:g1    domain2:g1			domain3:g2
	//		   /      \    /      \					|
	//	 domain1:u1    *:g1     domain2:u2		domain3:u4
	// 					|
	// 				   *:u3
	testDomainRole(t, rm, "u1", "g1", "domain1", true)
	testDomainRole(t, rm, "u2", "g1", "domain1", false)
	testDomainRole(t, rm, "u2", "g1", "domain2", true)
	testDomainRole(t, rm, "u3", "g1", "domain1", true)
	testDomainRole(t, rm, "u3", "g1", "domain2", true)
	testDomainRole(t, rm, "u1", "g2", "domain1", false)
	testDomainRole(t, rm, "u4", "g2", "domain3", true)
	testDomainRole(t, rm, "u3", "g2", "domain3", false)

	testPrintRolesWithDomain(t, rm, "u3", "domain1", []string{"g1"})
	testPrintRolesWithDomain(t, rm, "u1", "domain1", []string{"g1"})
	testPrintRolesWithDomain(t, rm, "u3", "domain2", []string{"g1"})
	testPrintRolesWithDomain(t, rm, "u1", "domain2", []string{})
	testPrintRolesWithDomain(t, rm, "u4", "domain3", []string{"g2"})
}

func TestAllMatchingFunc(t *testing.T) {
	rm := NewRoleManager(10)
	rm.(*RoleManager).AddMatchingFunc("keyMatch2", util.KeyMatch2)
	rm.(*RoleManager).AddDomainMatchingFunc("keyMatch2", util.KeyMatch2)

	_ = rm.AddLink("/book/:id", "book_group", "*")
	// Current role inheritance tree after deleting the links:
	//  		*:book_group
	//				|
	// 			*:/book/:id
	testDomainRole(t, rm, "/book/1", "book_group", "domain1", true)
	testDomainRole(t, rm, "/book/2", "book_group", "domain1", true)
}
