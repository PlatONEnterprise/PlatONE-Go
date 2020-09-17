package rest

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
