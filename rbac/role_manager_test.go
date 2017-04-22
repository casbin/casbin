package rbac

import (
	"github.com/hsluoyz/casbin/util"
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

func testPrintRoles(t *testing.T, rm *RoleManager, name string, res []string) {
	myRes := rm.GetRoles(name)
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
