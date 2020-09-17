package rest

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
		{"POST", "/node/components", testNodeAddBody, 200},
		{"DELETE", "/node/components/node0", testNodeUpdateBody, 200},

		{"GET", "/node/components", "", 200},
		{"GET", "/node/components?name=node0&endPoint=http://127.0.0.1:6791", "", 200},
		{"GET", "/node/components/statistic?name=node0", "", 200},

		{"GET", "/node/enode/deleted", "", 200}, // OMIT endPoint
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
