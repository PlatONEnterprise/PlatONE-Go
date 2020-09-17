package rest

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	txSender         = "0x063bc2e61696579cf4ad137fed8a7ced15501f73"
	testContractAddr = "0x942affd352030020d1d4e60160e99045f0c9cc21"
)

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
