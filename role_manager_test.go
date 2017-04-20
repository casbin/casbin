package casbin

import (
	"log"
	"testing"
)

func testRole(t *testing.T, rm *RoleManager, name1 string, name2 string, res bool) {
	myRes := rm.HasLink(name1, name2)
	log.Printf("%s, %s: %t", name1, name2, myRes)

	if myRes != res {
		t.Errorf("%s < %s: %t, supposed to be %t", name1, name2, !res, res)
	}
}

func testPrintRoles(rm *RoleManager, name string) {
	log.Print(name, ": ", rm.GetRoles(name))
}

func TestRole(t *testing.T) {
	rm := NewRoleManager(3)
	rm.AddLink("u1", "g1")
	rm.AddLink("u2", "g1")
	rm.AddLink("u3", "g2")
	rm.AddLink("u4", "g2")
	rm.AddLink("u4", "g3")
	rm.AddLink("g1", "g3")

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

	testPrintRoles(rm, "u1")
	testPrintRoles(rm, "u2")
	testPrintRoles(rm, "u3")
	testPrintRoles(rm, "u4")
	testPrintRoles(rm, "g1")
	testPrintRoles(rm, "g2")
	testPrintRoles(rm, "g3")
}
