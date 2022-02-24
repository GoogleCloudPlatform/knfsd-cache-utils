package mounts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitNFSDevice(t *testing.T) {
	var server, path string

	server, path = splitNFSDevice("example.com:/files/assets")
	assert.Equal(t, "example.com", server)
	assert.Equal(t, "/files/assets", path)

	// paths can contain colons, check that we only split the first colon
	server, path = splitNFSDevice("example.com:/files:assets")
	assert.Equal(t, "example.com", server)
	assert.Equal(t, "/files:assets", path)

	server, path = splitNFSDevice("")
	assert.Equal(t, "", server)
	assert.Equal(t, "", path)

	server, path = splitNFSDevice("foo")
	assert.Equal(t, "", server)
	assert.Equal(t, "foo", path)
}
