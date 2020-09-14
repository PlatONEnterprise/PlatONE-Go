package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	txSender         = "0x063bc2e61696579cf4ad137fed8a7ced15501f73"
	testContractAddr = "0xc52e02fb821334cd8a8145cafd7dd6ebafa634f8"
)

// ================== ACCOUNT ===========================
const (
	testNewAccountBody = "{\"\"}"
)

func TestAccountHandlers(t *testing.T) {
	testCase := []struct {
		method       string
		path         string
		body         string
		expectedCode int
	}{
		// cns
		{"POST", "/accounts", testCnsPostBody, 200},
	}

	router := genRestRouters()

	for _, data := range testCase {
		w := httptest.NewRecorder()

		param := make(url.Values)
		param.Set("passphrase", "123456")
		param.Set("privatekey", "")

		body := bytes.NewBufferString(param.Encode())
		req, _ := http.NewRequest(data.method, data.path, body)
		req.Header.Set("content-type", "application/x-www-form-urlencoded")

		router.ServeHTTP(w, req)

		assert.Equal(t, data.expectedCode, w.Code)
		t.Log(w.Body)
	}
}

// ================== Contract Name Service =========================
const (
	testCnsPostBody = "{\"tx\":{\"from\": \"" + txSender + "\", \"gas\":\"0x10\"}," +
		"\"contract\":{\"data\": {\"name\": \"tofu\", \"version\": \"0.0.0.1\", \"address\": \"" + testContractAddr + "\"},\"interpreter\": \"wasm\"}," +
		"\"rpc\":{\"endPoint\": \"http://127.0.0.1:6791\"}}"
)

func TestCnsHandlers(t *testing.T) {
	testCase := []struct {
		method       string
		path         string
		body         string
		expectedCode int
	}{
		// cns
		{"POST", "/cns/components", testCnsPostBody, 200},
		{"GET", "/cns/components?name=tofu&endPoint=http://127.0.0.1:6791", "", 200},
		{"GET", "/cns/components?page-num=1&page-size=2", "", 200},
		{"GET", "/cns/components?page-size=2", "", 200},
		{"GET", "/cns/components", "", 200},
		{"GET", "/cns/components/state?name=tofu&endPoint=http://127.0.0.1:6791", "", 200},

		{"PUT", "/cns/mappings/tofu", testCnsPostBody, 200},
		{"GET", "/cns/mappings/tofu?version=0.0.0.1&endPoint=http://127.0.0.1:6791", "", 200},

		// error cases
		{"GET", "/cns/components/state?name=tofu&address=" + testContractAddr + "&endPoint=http://127.0.0.1:6791", "", 400},
		{"GET", "/cns/components?name=@tofu", "", 400},
		{"GET", "/cns/components?origin=0x0023", "", 400},
		{"GET", "/cns/components?name=0&page-size=2", "", 400},
	}

	router := genRestRouters()

	for _, data := range testCase {
		w := httptest.NewRecorder()
		body := bytes.NewBufferString(data.body)
		req, _ := http.NewRequest(data.method, data.path, body)
		req.Header.Set("content-type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, data.expectedCode, w.Code)
		t.Log(w.Body)
	}
}

// ================== Node management =========================

const (
	testNodeAddBody = "{\"tx\":{\"from\": \"" + txSender + "\", \"gas\":\"0x10\"}," +
		"\"contract\":{\"data\": {\"info\": {\"name\":\"node0\",\"status\":1,\"internalIP\":\"127.0.0.1\",\"publicKey\":\"64a684197dbc77b69f418c511e55adb7a4a532a88d25d8e9d34667141d53790b5ff84ed385e35ade60ea9e610b3ac54499119fd1a9bf1344d319aeceadcb5bb7\",\"p2pPort\":8888}},\"interpreter\": \"wasm\"}," +
		"\"rpc\":{\"endPoint\": \"http://127.0.0.1:6791\"}}"

	testNodeUpdateBody = "{\"tx\":{\"from\": \"" + txSender + "\", \"gas\":\"0x10\"}," +
		"\"contract\":{\"data\": {\"info\": {\"status\":2}},\"interpreter\": \"wasm\"}," +
		"\"rpc\":{\"endPoint\": \"http://127.0.0.1:6791\"}}"
)

func TestNodeHandlers(t *testing.T) {
	testCase := []struct {
		method       string
		path         string
		body         string
		expectedCode int
	}{
		// node
		/// {"POST", "/node/components", testNodeAddBody, 200},
		/// {"DELETE", "/node/components/node0", testNodeUpdateBody, 200},

		{"GET", "/node/components", "", 200},
		{"GET", "/node/components?name=node0&endPoint=http://127.0.0.1:6791", "", 200},
		{"GET", "/node/components/statistic?name=node0", "", 200},

		/// {"GET", "/node/enode/deleted", "", 200}, // OMIT endPoint
	}

	router := genRestRouters()

	for _, data := range testCase {
		w := httptest.NewRecorder()
		body := bytes.NewBufferString(data.body)
		req, _ := http.NewRequest(data.method, data.path, body)
		req.Header.Set("content-type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, data.expectedCode, w.Code)
		t.Log(w.Body)
	}
}

// ================== system configuration =========================

const (
	testGasLimitBody = "{\"tx\":{\"from\": \"" + txSender + "\", \"gas\":\"0x10\"}," +
		"\"contract\":{\"data\":{\"blockGasLimit\":\"2000000000\"}}," +
		"\"rpc\":{\"endPoint\": \"http://127.0.0.1:6791\"}}"

	testGasLimitErrBody = "{\"tx\":{\"from\": \"" + txSender + "\", \"gas\":\"0x10\"}," +
		"\"contract\":{\"data\":{\"blockGasLimit\":\"1000000000\"}}," +
		"\"rpc\":{\"endPoint\": \"http://127.0.0.1:6791\"}}"

	testSysParamBody = "{\"tx\":{\"from\": \"" + txSender + "\", \"gas\":\"0x10\"}," +
		"\"contract\":{\"data\":{\"sysParam\":\"1\"}}," +
		"\"rpc\":{\"endPoint\": \"http://127.0.0.1:6791\"}}"

	testSysParamErrBody = "{\"tx\":{\"from\": \"" + txSender + "\", \"gas\":\"0x10\"}," +
		"\"contract\":{\"data\":{\"sysParam\":\"3\"}}," +
		"\"rpc\":{\"endPoint\": \"http://127.0.0.1:6791\"}}"

	testGasNameBody = "{\"tx\":{\"from\": \"" + txSender + "\", \"gas\":\"0x10\"}," +
		"\"contract\":{\"data\":{\"contractName\":\"tofu\"}}," +
		"\"rpc\":{\"endPoint\": \"http://127.0.0.1:6791\"}}"

	testGasNameErrBody = "{\"tx\":{\"from\": \"" + txSender + "\", \"gas\":\"0x10\"}," +
		"\"contract\":{\"data\":{\"contractName\":\"@alice\"}}," +
		"\"rpc\":{\"endPoint\": \"http://127.0.0.1:6791\"}}"
)

func TestSysConfigHandlers(t *testing.T) {
	testCase := []struct {
		method       string
		path         string
		body         string
		expectedCode int
	}{
		// sys
		{"PUT", "/sysConfig/block-gas-limit", testGasLimitBody, 200},
		{"GET", "/sysConfig/block-gas-limit", "", 200},

		{"PUT", "/sysConfig/is-produce-empty-block", testSysParamBody, 200},
		{"GET", "/sysConfig/is-produce-empty-block", "", 200},

		{"PUT", "/sysConfig/gas-contract-name", testGasNameBody, 200},
		{"GET", "/sysConfig/gas-contract-name", "", 200},

		// invalid
		{"PUT", "/sysConfig/block-gas-limit", testGasLimitErrBody, 400},
		{"PUT", "/sysConfig/is-produce-empty-block", testSysParamErrBody, 400},
		{"PUT", "/sysConfig/gas-contract-name", testGasNameErrBody, 400},
	}

	router := genRestRouters()

	for _, data := range testCase {
		w := httptest.NewRecorder()
		body := bytes.NewBufferString(data.body)
		req, _ := http.NewRequest(data.method, data.path, body)
		req.Header.Set("content-type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, data.expectedCode, w.Code)
		t.Log(w.Body)
	}
}

// ================== Fire Wall =========================

const (
	testFwNewRuleBody = "{\"tx\":{\"from\": \"" + txSender + "\", \"gas\":\"0x10\"}," +
		"\"contract\":{\"data\":{\"action\":\"ACCEPT\", \"rules\":\"0x5585cc9e9dc2b383fcbbbfd758744ada62427202:*|*:funcName2|0x5585cc9e9dc2b383fcbbbfd758744ada62427201:funcName3\"}}," +
		"\"rpc\":{\"endPoint\": \"http://127.0.0.1:6791\"}}"

	testFwClearRuleBody = "{\"tx\":{\"from\": \"" + txSender + "\", \"gas\":\"0x10\"}," +
		"\"contract\":{\"data\":{\"action\":\"ACCEPT\"}}," +
		"\"rpc\":{\"endPoint\": \"http://127.0.0.1:6791\"}}"

	testFwDeleteRuleBody = "{\"tx\":{\"from\": \"" + txSender + "\", \"gas\":\"0x10\"}," +
		"\"contract\":{\"data\":{\"action\":\"ACCEPT\", \"rules\":\"0x5585cc9e9dc2b383fcbbbfd758744ada62427201:funcName3\"}}," +
		"\"rpc\":{\"endPoint\": \"http://127.0.0.1:6791\"}}"

	testFwStatusBody = "{\"tx\":{\"from\": \"" + txSender + "\", \"gas\":\"0x10\"}," +
		"\"rpc\":{\"endPoint\": \"http://127.0.0.1:6791\"}}"

	testFwStatusErrBody = "{\"tx\":{\"from\": \"" + txSender + "\", \"gas\":\"0x10\"}," +
		"\"contract\":{\"data\":{\"address\":\"0x1000000000000000000000000000000000000001\"}}," +
		"\"rpc\":{\"endPoint\": \"http://127.0.0.1:6791\"}}"

	testFwOffBody = "{\"tx\":{\"from\": \"" + txSender + "\", \"gas\":\"0x10\"}," +
		"\"contract\":{\"data\":{\"status\":\"false\"}}," +
		"\"rpc\":{\"endPoint\": \"http://127.0.0.1:6791\"}}"
)

func TestFwHandlers(t *testing.T) {
	testCase := []struct {
		method       string
		path         string
		body         string
		expectedCode int
	}{
		{"PUT", "/fw/" + testContractAddr + "/on", testFwStatusBody, 200},
		{"POST", "/fw/" + testContractAddr + "/lists", testFwNewRuleBody, 200},
		{"PATCH", "/fw/" + testContractAddr + "/lists", testFwDeleteRuleBody, 200},
		{"GET", "/fw/" + testContractAddr, "", 200},

		{"DELETE", "/fw/" + testContractAddr + "/lists", testFwClearRuleBody, 200},
		{"PUT", "/fw/" + testContractAddr + "/off", testFwStatusBody, 200},
		{"GET", "/fw/" + testContractAddr, "", 200},

		// invalid
		{"PUT", "/fw/" + testContractAddr + "/on", testFwStatusErrBody, 200},
	}

	router := genRestRouters()

	for _, data := range testCase {
		w := httptest.NewRecorder()
		body := bytes.NewBufferString(data.body)
		req, _ := http.NewRequest(data.method, data.path, body)
		req.Header.Set("content-type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, data.expectedCode, w.Code)
		t.Log(w.Body)
	}
}

// ================== Role management =========================

const (
	testRoleBody = "{\"tx\":{\"from\": \"" + txSender + "\", \"gas\":\"0x10\"}," +
		"\"contract\":{\"data\":{\"address\":\"" + txSender + "\"}}," +
		"\"rpc\":{\"endPoint\": \"http://127.0.0.1:6791\"}}"

	testNullBody = "{\"tx\":{\"from\": \"" + txSender + "\", \"gas\":\"0x10\"}," +
		"\"rpc\":{\"endPoint\": \"http://127.0.0.1:6791\"}}"
)

func TestRoleHandlers(t *testing.T) {
	testCase := []struct {
		method       string
		path         string
		body         string
		expectedCode int
	}{
		{"POST", "/role/role-lists/super-admin", testNullBody, 200},
		{"PUT", "/role/role-lists/super-admin", testRoleBody, 200},

		{"PATCH", "/role/role-lists/contract-deployer", testRoleBody, 200},
		{"PATCH", "/role/role-lists/chain-admin", testRoleBody, 200},
		{"DELETE", "/role/role-lists/contract-deployer", testRoleBody, 200},

		{"GET", "/role/user-lists/" + txSender, "", 200},
		{"GET", "/role/role-lists/chain-admin", "", 200},
	}

	router := genRestRouters()

	for _, data := range testCase {
		w := httptest.NewRecorder()
		body := bytes.NewBufferString(data.body)
		req, _ := http.NewRequest(data.method, data.path, body)
		req.Header.Set("content-type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, data.expectedCode, w.Code)
		t.Log(w.Body)
	}
}

// ================== Contract Execute ====================

const (
	testContractDeployWasmBody = "{\"tx\":{\"from\": \"" + txSender + "\", \"gas\":\"0x10\"}," +
		"\"contract\":{\"data\": {\"params\": \"()\"},\"interpreter\": \"wasm\"}," +
		"\"rpc\":{\"endPoint\": \"http://127.0.0.1:6791\"}}"

	testContractDeployEvmBody = "{\"tx\":{\"from\": \"" + txSender + "\", \"gas\":\"0x10\"}," +
		"\"contract\":{\"data\": {\"params\": \"()\"},\"interpreter\": \"evm\"}," +
		"\"rpc\":{\"endPoint\": \"http://127.0.0.1:6791\"}}"

	testContractExecBody = "{\"tx\":{\"from\": \"" + txSender + "\", \"gas\":\"0x10\"}," +
		"\"contract\":{\"method\":\"invokeNotify\",\"data\": {\"params\": \"(this is a test)\"},\"interpreter\": \"wasm\"}," +
		"\"rpc\":{\"endPoint\": \"http://127.0.0.1:6791\"}}"

	testContractExecRawBody = "{\"tx\":{\"from\": \"0xf1b043d71ef5484d960a6f369a057386264d8c4b\", \"gas\":\"0x10\"}," +
		"\"contract\":{\"method\":\"invokeNotify\",\"data\": {\"params\": \"(this is a test)\"},\"interpreter\": \"wasm\"}," +
		"\"rpc\":{\"endPoint\": \"http://127.0.0.1:6791\", \"passphrase\":\"123456\"}}"
)

type uploadFile struct {
	path string
	name string
}

func NewUploadFile(name, path string) *uploadFile {
	return &uploadFile{
		name: name,
		path: path,
	}
}

func TestContractHandlers(t *testing.T) {
	testCase := []struct {
		method       string
		path         string
		body         string
		abiPath      string
		codePath     string
		expectedCode int
	}{
		{
			"POST",
			"/contracts",
			testContractDeployEvmBody,
			"../test/test_case/sol/privacyToken.abi",
			"../test/test_case/sol/privacyToken.bin",
			200},
		{
			"POST",
			"/contracts",
			testContractDeployWasmBody,
			"../test/test_case/wasm/appDemo.cpp.abi.json",
			"../test/test_case/wasm/appDemo.wasm",
			200},
	}

	router := genRestRouters()

	for _, data := range testCase {
		w := httptest.NewRecorder()

		file1 := NewUploadFile("code", data.codePath)
		file2 := NewUploadFile("abi", data.abiPath)

		/// resParam1, _ := UnmarshalToMap(data.body)
		resParam1 := make(map[string]string)
		resParam1["info"] = data.body

		body, contentType, err := genMultiPartBody([]*uploadFile{file1, file2}, resParam1)
		if err != nil {
			t.Error(err)
		}

		// s, _ := ioutil.ReadAll(body)
		// fmt.Printf("%s", s)

		req, _ := http.NewRequest(data.method, data.path, body)
		req.Header.Set("content-type", contentType)

		router.ServeHTTP(w, req)

		assert.Equal(t, data.expectedCode, w.Code)
		t.Log(w.Body)
	}
}

func genMultiPartBody(files []*uploadFile, reqParams ...map[string]string) (io.Reader, string, error) {
	body := bytes.NewBufferString("")
	writer := multipart.NewWriter(body)

	// write file to body
	for _, file := range files {
		f, err := os.Open(file.path)
		if err != nil {
			return nil, "", err
		}

		part, err := writer.CreateFormFile(file.name, filepath.Base(file.path))
		if err != nil {
			return nil, "", err
		}

		_, err = io.Copy(part, f)
		if err != nil {
			return nil, "", err
		}

		err = f.Close()
		if err != nil {
			return nil, "", err
		}
	}

	// write key value to body
	for _, category := range reqParams {
		for k, v := range category {
			_ = writer.WriteField(k, v)
		}
	}

	_ = writer.Close()

	return body, writer.FormDataContentType(), nil
}

func UnmarshalToMap(str string) (map[string]string, error) {
	var m map[string]interface{}

	err := json.Unmarshal([]byte(str), &m)
	if err != nil {
		return nil, err
	}

	info := make(map[string]string, 0)

	for k, v := range m {
		if reflect.TypeOf(v).Kind() == reflect.Map {
			value := v.(map[string]interface{})
			v = Maploop(value)
		}

		info[k] = fmt.Sprintf("%v", v)
	}

	return info, nil
}

func Maploop(m map[string]interface{}) string {
	var temp = make(url.Values, 0)

	for k, v := range m {
		if reflect.TypeOf(v).Kind() == reflect.Map {
			tempValue := v.(map[string]interface{})
			v = Maploop(tempValue)
		}

		/// temp.Set(k, fmt.Sprintf("%v", v))
		temp[k] = append(temp[k], fmt.Sprintf("%v", v))
	}

	return temp.Encode()
}

func TestJsonMarshalMap(t *testing.T) {
	str := testContractExecBody

	m, err := UnmarshalToMap(str)
	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v", m)
}

func TestContractExecHandlers(t *testing.T) {
	testCase := []struct {
		method       string
		path         string
		body         string
		expectedCode int
	}{
		/// {"POST", "/contract/" + testContractAddr, testContractExecBody, 200},
		{"POST", "/contracts/" + testContractAddr, testContractExecRawBody, 200},
	}

	router := genRestRouters()

	for _, data := range testCase {
		w := httptest.NewRecorder()

		file := NewUploadFile("file", "../test/test_case/wasm/appDemo.cpp.abi.json")

		/*
			params := make(url.Values, 0)
			params["param"] = []string{"param1", "param2"}

			resParam1 := make(map[string]string)
			resParam1["param"] = params.Encode()*/

		resParam1 := make(map[string]string)
		resParam1["info"] = data.body

		body, contentType, err := genMultiPartBody([]*uploadFile{file}, resParam1)
		if err != nil {
			t.Error(err)
		}

		req, _ := http.NewRequest(data.method, data.path, body)
		req.Header.Set("content-type", contentType)

		router.ServeHTTP(w, req)

		assert.Equal(t, data.expectedCode, w.Code)
		t.Log(w.Body)
	}
}
