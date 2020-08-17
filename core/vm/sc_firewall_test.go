package vm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	fwTestAddr1 = "0x0000000000000000000000000000000000000123"
	fwTestAddr2 = "0x0000000000000000000000000000000000000456"
	fwTestErr   = "0x00000000000000000000000000000000000002"
)

func TestConvertToFwElem(t *testing.T) {
	var err error
	testCases := []struct {
		rule     string
		expected error
	}{
		{fwTestAddr1 + ":func1|" + fwTestAddr2 + ":func2", nil},
		{fwTestAddr1 + "func1", ErrFwRule},
		{fwTestErr + ":func1", ErrFwRuleAddr},
		{fwTestAddr1 + ":*", nil},
	}

	for _, data := range testCases {
		_, err = convertToFwElem(data.rule)
		assert.Equal(t, data.expected, err, "bug in convertToFwElem")
	}

}
