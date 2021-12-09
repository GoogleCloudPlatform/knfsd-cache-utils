package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"netapp-exports/internal/testcert"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var teapot http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTeapot)
}

func TestTransport(t *testing.T) {
	server := newTLSServer(testcert.LocalhostCert, teapot)
	defer server.Close()

	tls := &TLSConfig{
		CACertificate: string(testcert.LocalhostCert),
	}
	transport, err := tls.transport()
	require.NoError(t, err)

	client := &http.Client{
		Transport: transport,
	}

	resp, err := client.Get(server.URL)
	defer safeClose(resp)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusTeapot, statusCode(resp))
}

func TestTransportHttpRequiresInsecure(t *testing.T) {
	query := func(insecure bool) (int, error) {
		tls := &TLSConfig{
			insecure: insecure,
		}

		server := httptest.NewServer(teapot)
		defer server.Close()

		client := httpClient(tls)
		resp, err := client.Get(server.URL)
		defer safeClose(resp)

		return statusCode(resp), err
	}

	t.Run("insecure false", func(t *testing.T) {
		status, err := query(false)

		assert.Equal(t, 0, status)
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "HTTP not permitted unless insecure specified")
		}
	})

	t.Run("insecure true", func(t *testing.T) {
		status, err := query(true)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusTeapot, status)
	})
}

func TestTransportInsecureIgnoresServerCertificate(t *testing.T) {
	query := func(insecure bool) (int, error) {
		tls := &TLSConfig{
			insecure: insecure,
		}

		server := newTLSServer(testcert.LocalhostCert, teapot)
		defer server.Close()

		client := httpClient(tls)
		resp, err := client.Get(server.URL)
		defer safeClose(resp)

		return statusCode(resp), err
	}

	t.Run("insecure false", func(t *testing.T) {
		status, err := query(false)

		assert.Equal(t, 0, status)
		assert.Error(t, err)

		if err != nil {
			assert.Contains(t, err.Error(), "x509: certificate signed by unknown authority")
		}
	})

	t.Run("insecure true", func(t *testing.T) {
		status, err := query(true)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusTeapot, status)
	})
}

func TestTransportCommonName(t *testing.T) {
	type testCase struct {
		name            string
		host            string
		allowCommonName bool
		expectedError   string
	}

	tests := []testCase{
		// check certificate works when using the correct host
		{"using common name", "localhost", true, ""},

		// check that it still correctly detects when the host does not match the common name
		// i.e. verify that it is still checking and doesn't work simply because all validation is disabled
		{"common name does not match", "127.0.0.1", true, "x509: certificate is valid for localhost"},

		// check that common name is not allowed when AllowCommonName is false
		{"common name not allowed", "localhost", false, "x509: certificate relies on legacy Common Name field"},

		// TODO: check other forms of invalid certificate such as expired date, wrong key usage, etc
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := newTLSServer(testcert.CommonNameCert, teapot)
			defer server.Close()

			tls := &TLSConfig{
				CACertificate:   string(testcert.CommonNameCert),
				AllowCommonName: tc.allowCommonName,
			}

			client := httpClient(tls)
			url := fmt.Sprintf("https://%s:%d", tc.host, port(server))
			resp, err := client.Get(url)
			defer safeClose(resp)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, http.StatusTeapot, statusCode(resp))
			} else {
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, err.Error(), tc.expectedError)
				}
				assert.Nil(t, resp)
			}
		})
	}
}

func newTLSServer(certPEMBlock []byte, handler http.Handler) *httptest.Server {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	// Suppress spurious "remote error: tls: bad certificate" warnings
	server.Config.ErrorLog = log.New(io.Discard, "", 0)

	cert, err := tls.X509KeyPair(certPEMBlock, testcert.PrivateKey)
	if err != nil {
		panic(fmt.Sprintf("newTLSServer: %v", err))
	}

	if server.TLS == nil {
		server.TLS = &tls.Config{}
	}

	server.TLS.Certificates = []tls.Certificate{cert}

	server.StartTLS()
	return server
}

func safeClose(resp *http.Response) error {
	if resp != nil && resp.Body != nil {
		return resp.Body.Close()
	}
	return nil
}

func httpClient(tls *TLSConfig) *http.Client {
	transport, err := tls.transport()
	if err != nil {
		panic(fmt.Sprintf("httpClient: %v", err))
	}
	return &http.Client{
		Transport: transport,
	}
}

func port(s *httptest.Server) int {
	return s.Listener.Addr().(*net.TCPAddr).Port
}

func statusCode(resp *http.Response) int {
	if resp == nil {
		return 0
	}
	return resp.StatusCode
}
