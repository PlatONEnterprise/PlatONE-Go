package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsRoleMatch(t *testing.T) {
	testCases := []struct {
		role     string
		expected bool
	}{
		{"chainCreator", true},
		{"", false},
		{"chain Creator", false},
	}

	for _, data := range testCases {
		result := IsRoleMatch(data.role)
		assert.Equal(t, data.expected, result, "error")
	}
}

func TestPrintStrArray(t *testing.T) {
	strArray := []string{"abs", "is a test"}
	t.Logf("%s", strArray)
}
