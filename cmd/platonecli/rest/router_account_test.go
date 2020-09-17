package rest

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ================== ACCOUNT ===========================
func TestAccountHandlers(t *testing.T) {
	testCase := []struct {
		method       string
		path         string
		body         string
		expectedCode int
	}{
		// cns
		{"POST", "/accounts", "", 200},
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
