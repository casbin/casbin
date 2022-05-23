package fileadapter

import "testing"

func TestFilterWordsMCC(t *testing.T) {
	assert.Equal(t, filterWords([]string{"a"}, []string{"b"}), true)
	assert.Equal(t, filterWords([]string{"a", "b"}, []string{"c"}), true)
	assert.Equal(t, filterWords([]string{"a", "b"}, []string{"b"}), false)
	assert.Equal(t, filterWords([]string{"a", "b"}, []string{""}), false)
	assert.Equal(t, filterWords([]string{"a", ""}, []string{""}), false)
	assert.Equal(t, filterWords([]string{"a", "b"}, []string{}), false)
}


func TestFilterWordsDUPath(t *testing.T) {
	assert.Equal(t, filterWords([]string{"bac"}, []string{"cdefe"}), true)
	assert.Equal(t, filterWords([]string{"cde", "cde"}, []string{}), false)
	assert.Equal(t, filterWords([]string{"ce", "bf"}, []string{"gh"}), true)
}
