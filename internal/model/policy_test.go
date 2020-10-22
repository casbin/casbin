package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewPolicy(t *testing.T) {
	p := NewPolicy()
	assert.NotNil(t, p)
}

func TestPolicy_Index(t *testing.T) {
	p := NewPolicy()
	rule := []string{"alice", "data1", "read"}
	p.addIndex(rule, 0)
	assert.Equal(t, p.hasIndex(rule), true)
	p.removeIndex(rule)
	assert.Equal(t, p.hasIndex(rule), false)
}

func TestToIndexKey(t *testing.T) {
	key := toIndexKey([]string{"alice", "data", "read"})
	assert.NotEmpty(t, key)
}

func TestPolicy_AddPolicy(t *testing.T) {
	type testcase struct {
		rule     []string
		expected bool
	}

	cases := []testcase{
		{
			rule:     []string{"alice", "data1", "read"},
			expected: true,
		}, {
			rule:     []string{"alice", "data2", "read"},
			expected: true,
		}, {
			rule:     []string{"bob", "data1", "write"},
			expected: true,
		}, {
			rule:     []string{"alice", "data1", "read"},
			expected: false,
		},
	}

	p := NewPolicy()
	for _, tc := range cases {
		ok := p.AddPolicy(tc.rule)
		assert.Equal(t, tc.expected, ok)
	}
}

func TestPolicy_AddPolicies(t *testing.T) {
	type testcase struct {
		rule     [][]string
		expected [][]string
	}

	cases := []testcase{
		{
			rule:     [][]string{{"alice", "data1", "read"}, {"alice", "data2", "read"}},
			expected: [][]string{{"alice", "data1", "read"}, {"alice", "data2", "read"}},
		}, {
			rule:     [][]string{{"alice", "data1", "read"}, {"alice", "data2", "write"}},
			expected: [][]string{{"alice", "data2", "write"}},
		},
	}

	p := NewPolicy()
	for _, tc := range cases {
		ok := p.AddPolicies(tc.rule)
		assert.Equal(t, tc.expected, ok)
	}
}

func TestPolicy_AddAndRemovePolicy(t *testing.T) {
	p := NewPolicy()

	cases := [][]string{
		{"alice", "data1", "read"},
		{"alice", "data2", "read"},
		{"alice", "data3", "read"},
	}

	for _, tc := range cases {
		ok := p.AddPolicy(tc)
		assert.Equal(t, true, ok)
	}

	for _, tc := range cases {
		ok := p.RemovePolicy(tc)
		assert.Equal(t, true, ok)
	}
}

func TestPolicy_AddAndRemovePolicies(t *testing.T) {
	p := NewPolicy()

	assert.Empty(t, p.RemovePolicies(nil))

	assert.Equal(t,
		p.AddPolicies([][]string{{"alice", "data1", "read"}, {"alice", "data2", "read"}, {"bob", "data1", "read"}}),
		[][]string{{"alice", "data1", "read"}, {"alice", "data2", "read"}, {"bob", "data1", "read"}},
	)

	assert.Equal(t,
		p.RemovePolicies([][]string{{"alice", "data1", "read"}}),
		[][]string{{"alice", "data1", "read"}},
	)

	assert.Equal(t, p.RemoveFilteredPolicy(1, "data2"), [][]string{{"alice", "data2", "read"}})
}

func TestPolicy_GetPolicy(t *testing.T) {
	p := NewPolicy()

	assert.Empty(t, p.GetPolicy())

	p.AddPolicy([]string{"alice", "data1", "read"})
	assert.Equal(t, p.GetPolicy(), [][]string{{"alice", "data1", "read"}})

	p.RemovePolicy([]string{"alice", "data1", "read"})
	assert.Empty(t, p.GetPolicy())

	assert.Empty(t, p.FilterExistsPolicy(nil))

	p.AddPolicy([]string{"alice", "data1", "read"})
	p.AddPolicy([]string{"alice", "data2", "read"})
	p.AddPolicy([]string{"bob", "data2", "read"})

	assert.Equal(t, p.FilterExistsPolicy(
		[][]string{{"alice", "data2", "read"}, {"alice", "data3", "read"}}),
		[][]string{{"alice", "data2", "read"}},
	)

	assert.Equal(t, p.FilterNotExistsPolicy(
		[][]string{{"alice", "data2", "read"}, {"alice", "data3", "read"}}),
		[][]string{{"alice", "data3", "read"}},
	)

	assert.Equal(t, p.GetValuesForFieldInPolicy(0), []string{"alice", "bob"})

	assert.Equal(t,
		p.GetFilteredPolicy(1, "data2"),
		[][]string{
			{"alice", "data2", "read"},
			{"bob", "data2", "read"},
		},
	)
}

func TestPolicy_ClearPolicy(t *testing.T) {
	p := NewPolicy()

	p.AddPolicy([]string{"alice", "data1", "read"})
	p.AddPolicy([]string{"alice", "data2", "read"})

	assert.Equal(t, p.GetPolicy(), [][]string{
		{"alice", "data1", "read"},
		{"alice", "data2", "read"},
	})

	p.ClearPolicy()
	assert.Empty(t, p.GetPolicy())
}

func TestPolicy_HasPolicy(t *testing.T) {
	p := NewPolicy()

	p.AddPolicy([]string{"alice", "data1", "read"})
	assert.Equal(t, p.HasPolicy([]string{"alice", "data1", "read"}), true)

	p.RemovePolicy([]string{"alice", "data1", "read"})
	assert.Equal(t, p.HasPolicy([]string{"alice", "data1", "read"}), false)

	assert.Equal(t, p.HasPolicy([]string{"bob", "data1", "read"}), false)
}
