package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func api(server *httptest.Server) *API {
	return &API{
		Client:  server.Client(),
		BaseURL: server.URL + "/api/v1/",
	}
}

func TestFetchPage(t *testing.T) {
	tests := map[string]VolumePathList{
		"testdata/responses/empty.json": {},

		"testdata/responses/first_page.json": {
			Paths:    []string{"/", "/archive"},
			NextPage: "/api/v1/storage/volumes?start.uuid=4ababbd0-2da2-11ec-86fb-ebf808c63a47&nas.path=!null&fields=nas.path&max_records=100",
		},

		"testdata/responses/last_page.json": {
			Paths: []string{"/assets"},
		},

		"testdata/responses/partial.json": {
			Paths: []string{"/foo", "/bar"},
		},

		"testdata/responses/empty_strings.json": {
			Paths: []string{"/foo", "/bar"},
		},

		"testdata/responses/null_strings.json": {
			Paths: []string{"/foo", "/bar"},
		},
	}

	for file, expected := range tests {
		t.Run(file, func(t *testing.T) {
			json, err := os.ReadFile(file)
			require.NoError(t, err)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/storage/volumes?nas.path=!null&fields=nas.path&max_records=100", r.RequestURI)
				w.Write(json)
			}))
			defer server.Close()

			if expected.NextPage != "" {
				expected.NextPage = server.URL + expected.NextPage
			}

			api := api(server)
			page, err := api.FetchPage()

			assert.NoError(t, err)
			assert.Equal(t, expected, page)
		})
	}
}

func TestFetchPage_Auth(t *testing.T) {
	invoked := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "admin", user)
		assert.Equal(t, "test", pass)
		invoked = true
		w.WriteHeader(401)
	}))
	defer server.Close()

	api := api(server)
	api.User = "admin"
	api.Password = "test"
	api.FetchPage()

	assert.True(t, invoked)
}

func TestFetchPage_HttpError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer server.Close()

	api := api(server)
	page, err := api.FetchPage()

	assert.Error(t, err)
	assert.Equal(t, VolumePathList{}, page)
}

func TestFetchPage_InvalidJson(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><body>Something went wrong</body></html>"))
	}))
	defer server.Close()

	api := api(server)
	page, err := api.FetchPage()

	assert.Error(t, err)
	assert.Equal(t, VolumePathList{}, page)
}

func TestFetchAll(t *testing.T) {
	type response struct {
		expectedURI string
		jsonFile    string
	}
	responses := make(chan response)
	go func() {
		responses <- response{
			"/api/v1/storage/volumes?nas.path=!null&fields=nas.path&max_records=100",
			"testdata/responses/first_page.json",
		}
		responses <- response{
			"/api/v1/storage/volumes?start.uuid=4ababbd0-2da2-11ec-86fb-ebf808c63a47&nas.path=!null&fields=nas.path&max_records=100",
			"testdata/responses/last_page.json",
		}
		close(responses)
	}()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next, ok := <-responses
		if !ok {
			assert.Fail(t, "Too many requests")
			return
		}

		assert.Equal(t, next.expectedURI, r.RequestURI)
		json, err := os.ReadFile(next.jsonFile)
		require.NoError(t, err)
		w.Write(json)
	}))
	defer server.Close()

	api := api(server)
	paths, err := api.FetchAll()

	assert.NoError(t, err)
	assert.Equal(t, []string{"/", "/archive", "/assets"}, paths)

	_, ok := <-responses
	if ok {
		assert.Fail(t, "Not enough requests")
		for range responses {
		}
	}
}
