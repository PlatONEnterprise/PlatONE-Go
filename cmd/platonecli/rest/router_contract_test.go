package rest

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
