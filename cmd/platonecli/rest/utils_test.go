package rest

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Address string `json:"address"`
}

func newTestInfo(name, version, address string) *testInfo {
	return &testInfo{
		Name:    name,
		Version: version,
		Address: address,
	}
}

func TestGetDataParams(t *testing.T) {
	testCases := []struct {
		Info     interface{}
		expected bool
	}{
		// correct cases
		{newTestInfo("", "", ""), true},
		{newTestInfo("tofu", "", ""), true},
		{newTestInfo("", "0.0.0.1", ""), true},
		{newTestInfo("", "", "0x9ccf0b561c9142d3a771ce2131db8bc9fba61f6f"), true},
		{newTestInfo("tofu", "0.0.0.1", "0x9ccf0b561c9142d3a771ce2131db8bc9fba61f6f"), true},

		// error cases
		{newTestInfo("", "0.0.0.x", ""), false},
		{newTestInfo("", "", "0x231"), false},

		// other cases
		{*newTestInfo("tofu", "0.0.0.1", "0x9ccf0b561c9142d3a771ce2131db8bc9fba61f6f"), true},
		{[]string{"tofu"}, true},
	}

	for _, data := range testCases {
		res, err := getDataParams(data.Info)
		if err != nil {
			t.Errorf("error: %v", err)
		}

		valid := paramsCheck(data.Info)
		if valid != data.expected {
			t.Error(res, "error: something wrong")
		} else {
			t.Log(res, valid)
		}
	}

}

func TestJsonUnmarshal(t *testing.T) {
	testCase := "{\"status\":2}"

	testStruct := &struct {
		Status string `json:"status"`
	}{}

	err := json.Unmarshal([]byte(testCase), testStruct)
	if err != nil {
		t.Error(err)
	}

	t.Logf("%v", testStruct)
}

func TestUrlParamConvert(t *testing.T) {
	testCase := []struct {
		str           string
		expectedStr   string
		expectedStrV2 string
	}{
		{"is-approve-deployed-contract", "isApproveDeployedContract", "IS_APPROVE_DEPLOYED_CONTRACT"},
	}

	for _, data := range testCase {
		res := UrlParamConvert(data.str)
		resV2 := UrlParamConvertV2(data.str)

		assert.Equal(t, res, data.expectedStr)
		assert.Equal(t, resV2, data.expectedStrV2)

	}
}

type result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// 34: "
// 123: {
// 92: \
func TestUnmarshal(t *testing.T) {
	var i interface{}
	str := `"{\"code\":1,\"msg\":\"[CNS] the matching list is empty\",\"data\":[]}"`
	str2 := `{\"code\":1,\"msg\":\"[CNS] the matching list is empty\",\"data\":[]}`
	str3 := `{"code":1,"msg":"[CNS] the matching list is empty","data":[]}`
	t.Log("1:", []byte(str), str)
	t.Log("2:", []byte(str2), str2)
	t.Log("3:", []byte(str3), str3)

	err := json.Unmarshal([]byte(str), &i)
	if err != nil {
		t.Error(err)
	}

	resBytes, _ := json.Marshal(i)
	t.Log("t1:", resBytes)

	var test = &result{
		Code: 1,
		Msg:  "[CNS] the matching list is empty",
		Data: nil,
	}

	resBytes, _ = json.Marshal(test)
	t.Log("t3.1:", resBytes)

	resBytes, _ = json.Marshal(str3)
	t.Log("t3.2:", resBytes)
}
