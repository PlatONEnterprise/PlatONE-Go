package utils

import (
	"testing"
)

const (
	TEST_ACCOUNT = "0x60ceca9c1290ee56b98d4e160ef0453f7c40d219"
)

func TestParamValid(t *testing.T) {
	testCase := []struct {
		param     string
		paramName string
	}{
		{"*", "fw"},
		{TEST_ACCOUNT, "fw"},
		{TEST_ACCOUNT, "contract"},
		{"Alice_02", "contract"},
		//{"Alice.bob", "contract"},
		{"accept", "action"},
		//{"xxx", "action"},
		{"127.0.0.1:6791", "url"},
		{"127.0.0.1", "externalIP"},
		{"[\"nodeAdmin \"]", "roles"},
		{"fd.deng@wxblockchain.com", "email"},
		{"13240283946", "mobile"},
		{"0.0.0.1", "version"},
		{"-123", "num"},
		{"+13", "num"},
		{"12459234", "num"},
		// {"+-123", "num"},
	}

	for i, data := range testCase {
		ParamValid(data.param, data.paramName)
		t.Logf("case %d: the %s \"%s\" is valid\n", i, data.paramName, data.param)

	}

}
