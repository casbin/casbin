package casbin

import (
	"fmt"
	"testing"
)

func BenchmarkRoleManagerSmall(b *testing.B) {
	e, _ := NewEnforcer("examples/rbac_model.conf", false)
	// Do not rebuild the role inheritance relations for every AddGroupingPolicy() call.
	e.EnableAutoBuildRoleLinks(false)

	// 100 roles, 10 resources.
	pPolicies := make([][]string, 0)
	for i := 0; i < 100; i++ {
		pPolicies = append(pPolicies, []string{fmt.Sprintf("group%d", i), fmt.Sprintf("data%d", i/10), "read"})
	}

	_, err := e.AddPolicies(pPolicies)
	if err != nil {
		b.Fatal(err)
	}

	// 1000 users.
	gPolicies := make([][]string, 0)
	for i := 0; i < 1000; i++ {
		gPolicies = append(gPolicies, []string{fmt.Sprintf("user%d", i), fmt.Sprintf("group%d", i/10)})
	}

	_, err = e.AddGroupingPolicies(gPolicies)
	if err != nil {
		b.Fatal(err)
	}

	rm := e.GetRoleManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			_, _ = rm.HasLink("user501", fmt.Sprintf("group%d", j))
		}
	}
}

func BenchmarkRoleManagerMedium(b *testing.B) {
	e, _ := NewEnforcer("examples/rbac_model.conf", false)
	// Do not rebuild the role inheritance relations for every AddGroupingPolicy() call.
	e.EnableAutoBuildRoleLinks(false)

	// 1000 roles, 100 resources.
	pPolicies := make([][]string, 0)
	for i := 0; i < 1000; i++ {
		pPolicies = append(pPolicies, []string{fmt.Sprintf("group%d", i), fmt.Sprintf("data%d", i/10), "read"})
	}
	_, err := e.AddPolicies(pPolicies)
	if err != nil {
		b.Fatal(err)
	}

	// 10000 users.
	gPolicies := make([][]string, 0)
	for i := 0; i < 10000; i++ {
		gPolicies = append(gPolicies, []string{fmt.Sprintf("user%d", i), fmt.Sprintf("group%d", i/10)})
	}

	_, err = e.AddGroupingPolicies(gPolicies)
	if err != nil {
		b.Fatal(err)
	}

	err = e.BuildRoleLinks()
	if err != nil {
		b.Fatal(err)
	}

	rm := e.GetRoleManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 1000; j++ {
			_, _ = rm.HasLink("user501", fmt.Sprintf("group%d", j))
		}
	}
}

func BenchmarkRoleManagerLarge(b *testing.B) {
	e, _ := NewEnforcer("examples/rbac_model.conf", false)

	// 10000 roles, 1000 resources.
	pPolicies := make([][]string, 0)
	for i := 0; i < 10000; i++ {
		pPolicies = append(pPolicies, []string{fmt.Sprintf("group%d", i), fmt.Sprintf("data%d", i/10), "read"})
	}

	_, err := e.AddPolicies(pPolicies)
	if err != nil {
		b.Fatal(err)
	}

	// 100000 users.
	gPolicies := make([][]string, 0)
	for i := 0; i < 100000; i++ {
		gPolicies = append(gPolicies, []string{fmt.Sprintf("user%d", i), fmt.Sprintf("group%d", i/10)})
	}

	_, err = e.AddGroupingPolicies(gPolicies)
	if err != nil {
		b.Fatal(err)
	}

	rm := e.GetRoleManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 10000; j++ {
			_, _ = rm.HasLink("user501", fmt.Sprintf("group%d", j))
		}
	}
}
