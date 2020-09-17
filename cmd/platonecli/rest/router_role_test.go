package rest

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
