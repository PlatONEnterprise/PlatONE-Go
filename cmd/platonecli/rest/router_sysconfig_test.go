package rest

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
		/// {"PUT", "/sysConfig/block-gas-limit", testGasLimitBody, 200},
		/// {"GET", "/sysConfig/block-gas-limit", "", 200},

		/// {"PUT", "/sysConfig/is-produce-empty-block", testSysParamBody, 200},
		/// {"GET", "/sysConfig/is-produce-empty-block", "", 200},

		/// {"PUT", "/sysConfig/check-contract-deploy-permission", testSysParamBody, 200},
		/// {"GET", "/sysConfig/check-contract-deploy-permission", "", 200},

		/// {"PUT", "/sysConfig/gas-contract-name", testGasNameBody, 200},
		{"GET", "/sysConfig/gas-contract-name", "", 200},

		// invalid
		/// {"PUT", "/sysConfig/block-gas-limit", testGasLimitErrBody, 400},
		/// {"PUT", "/sysConfig/is-produce-empty-block", testSysParamErrBody, 400},
		/// {"PUT", "/sysConfig/gas-contract-name", testGasNameErrBody, 400},
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
