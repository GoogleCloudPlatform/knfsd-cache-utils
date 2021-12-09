package main

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfig(t *testing.T) {
	c := parseTestConfig(t, "basic.hcl")
	assert.Len(t, c.Servers, 3)

	t.Run("basic-attributes", func(t *testing.T) {
		s := findServer(t, c, "basic-attributes")
		assert.Equal(t, "https://10.0.0.2:8080", s.URL)
		assert.Equal(t, "nfs-proxy", s.User)
		assert.Equal(t, "secret", s.Password)

		// even though we did not set a TLS block, the TLS block should still
		// have a default value
		assert.NotNil(t, s.TLS)
		assert.Equal(t, "", s.TLS.CACertificate)
		assert.Equal(t, false, s.TLS.AllowCommonName)
		assert.Equal(t, false, s.TLS.insecure)
	})

	t.Run("tls", func(t *testing.T) {
		s := findServer(t, c, "tls")
		require.NotNil(t, s.TLS)
		assert.Equal(t, "NetApp CA Certificate", s.TLS.CACertificate)
		assert.Equal(t, true, s.TLS.AllowCommonName)
		assert.Equal(t, false, s.TLS.insecure)
	})

	t.Run("empty-tls", func(t *testing.T) {
		s := findServer(t, c, "empty-tls")
		require.NotNil(t, s.TLS)
		assert.Equal(t, "", s.TLS.CACertificate)
		assert.Equal(t, false, s.TLS.AllowCommonName)
		assert.Equal(t, false, s.TLS.insecure)
	})
}

func TestParseConfig_GCPSecret(t *testing.T) {
	c := parseTestConfig(t, "google_cloud_secret.hcl")
	assert.Len(t, c.Servers, 3)

	t.Run("all-attributes", func(t *testing.T) {
		s := findGCPSecret(t, c, "all-attributes")
		assert.Equal(t, "Service Account Key", s.ServiceAccountKey)
		assert.Equal(t, "example", s.Project)
		assert.Equal(t, "netapp-password", s.Name)
		assert.Equal(t, "latest", s.Version)
	})

	t.Run("minimal", func(t *testing.T) {
		s := findGCPSecret(t, c, "minimal")
		assert.Equal(t, "", s.ServiceAccountKey)
		assert.Equal(t, "", s.Project)
		assert.Equal(t, "netapp-password", s.Name)
		assert.Equal(t, "", s.Version)
	})

	t.Run("version-number", func(t *testing.T) {
		s := findGCPSecret(t, c, "numeric-version")
		assert.Equal(t, "netapp-password", s.Name)
		assert.Equal(t, "42", s.Version)
	})
}

func parseTestConfig(t *testing.T, name string) *Config {
	t.Helper()

	baseDir := "testdata/config"
	file := path.Join(baseDir, name)

	c, err := parseConfigFile(file)
	require.NoError(t, err)

	return c
}

func findServer(t *testing.T, c *Config, host string) *NetAppServer {
	t.Helper()
	for _, s := range c.Servers {
		if s.Host == host {
			return s
		}
	}
	require.Fail(t, "could not find server "+host)
	return nil
}

func findGCPSecret(t *testing.T, c *Config, name string) *GCPSecret {
	t.Helper()
	s := findServer(t, c, name)
	require.NotNil(t, s.SecurePassword)
	require.NotNil(t, s.SecurePassword.GCPSecret)
	return s.SecurePassword.GCPSecret
}
