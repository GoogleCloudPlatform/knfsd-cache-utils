package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListExports(t *testing.T) {
	json, err := os.ReadFile("testdata/responses/simple.json")
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(json)
	}))
	defer server.Close()

	s := &NetAppServer{
		Host:     "netapp.test",
		URL:      server.URL,
		User:     "test",
		Password: "test",
		TLS: &TLSConfig{
			insecure: true,
		},
	}

	err = s.validate()
	require.NoError(t, err)

	expected := strings.Join([]string{
		"netapp.test /",
		"netapp.test /archive",
		"netapp.test /assets",
		"",
	}, "\n")

	actual := new(strings.Builder)
	err = listExports(actual, s)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual.String())
}
