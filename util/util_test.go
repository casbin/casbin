package util

import (
	"log"
	"testing"
)

func testFixAttribute(t *testing.T, s string, res string) {
	myRes := FixAttribute(s)
	log.Printf("%s: %s", s, myRes)

	if myRes != res {
		t.Errorf("%s: %s, supposed to be %s", s, myRes, res)
	}
}

func TestFixAttribute(t *testing.T) {
	testFixAttribute(t, "r.sub.domain", "subAttr(r.sub, \"domain\")")
	testFixAttribute(t, "r.obj.role", "objAttr(r.obj, \"role\")")
	testFixAttribute(t, "r.act.url", "actAttr(r.act, \"url\")")
	testFixAttribute(t, "p.sub.domain", "subAttr(p.sub, \"domain\")")
	testFixAttribute(t, "p.obj.role", "objAttr(p.obj, \"role\")")
	testFixAttribute(t, "p.act.url", "actAttr(p.act, \"url\")")
}
