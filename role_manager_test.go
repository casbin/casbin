package casbin

import (
	"testing"
	"log"
)

func testRole(t *testing.T, rm *RoleManager, name1 string, name2 string, res bool) {
	my_res := rm.hasLink(name1, name2)
	log.Printf("%s, %s: %t", name1, name2, my_res)

	if my_res != res {
		t.Errorf("%s < %s: %t, supposed to be %t", name1, name2, !res, res)
	}
}

func TestRole(t *testing.T) {
	rm := newRoleManager(3)
	rm.addLink("u1", "g1")
	rm.addLink("u2", "g1")
	rm.addLink("u3", "g2")
	rm.addLink("g1", "g3")

	testRole(t, rm, "u1", "g1", true)
	testRole(t, rm, "u1", "g2", false)
	testRole(t, rm, "u1", "g3", true)
	testRole(t, rm, "u2", "g1", true)
	testRole(t, rm, "u2", "g2", false)
	testRole(t, rm, "u2", "g3", true)
	testRole(t, rm, "u3", "g1", false)
	testRole(t, rm, "u3", "g2", true)
	testRole(t, rm, "u3", "g3", false)
}
