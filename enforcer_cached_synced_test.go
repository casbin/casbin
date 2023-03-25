package casbin

import (
	"testing"
)

func testSyncEnforceCache(t *testing.T, e *SyncCachedEnforcer, sub string, obj interface{}, act string, res bool) {
	t.Helper()
	if myRes, _ := e.Enforce(sub, obj, act); myRes != res {
		t.Errorf("%s, %v, %s: %t, supposed to be %t", sub, obj, act, myRes, res)
	}
}

func TestSyncCache(t *testing.T) {
	e, _ := NewSyncedCachedEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")
	// The cache is enabled by default for NewCachedEnforcer.
	for i := 0; i < 20; i++ {
		go func() {
			testSyncEnforceCache(t, e, "alice", "data1", "write", false)
		}()
	}

	testSyncEnforceCache(t, e, "alice", "data1", "read", true)
	testSyncEnforceCache(t, e, "alice", "data1", "write", false)
	testSyncEnforceCache(t, e, "alice", "data2", "read", false)
	testSyncEnforceCache(t, e, "alice", "data2", "write", false)
	// The cache is enabled, calling RemovePolicy, LoadPolicy or RemovePolicies will
	// also operate cached items.
	_, _ = e.RemovePolicy("alice", "data1", "read")

	testSyncEnforceCache(t, e, "alice", "data1", "read", false)
	testSyncEnforceCache(t, e, "alice", "data1", "write", false)
	testSyncEnforceCache(t, e, "alice", "data2", "read", false)
	testSyncEnforceCache(t, e, "alice", "data2", "write", false)

	e, _ = NewSyncedCachedEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")

	testSyncEnforceCache(t, e, "alice", "data1", "read", true)
	testSyncEnforceCache(t, e, "bob", "data2", "write", true)
	testSyncEnforceCache(t, e, "alice", "data2", "read", true)
	testSyncEnforceCache(t, e, "alice", "data2", "write", true)

	_, _ = e.RemovePolicies([][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "write"},
	})

	testSyncEnforceCache(t, e, "alice", "data1", "read", false)
	testSyncEnforceCache(t, e, "bob", "data2", "write", false)
	testSyncEnforceCache(t, e, "alice", "data2", "read", true)
	testSyncEnforceCache(t, e, "alice", "data2", "write", true)
}
