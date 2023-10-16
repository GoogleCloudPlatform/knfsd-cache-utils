package main

import (
	"net/http"

	"golang.org/x/sys/unix"
)

type CacheUsageResponse struct {
	BytesTotal     uint64 `json:"bytesTotal"`
	BytesUsed      uint64 `json:"bytesUsed"`
	BytesFree      uint64 `json:"bytesFree"`
	BytesAvailable uint64 `json:"bytesAvailable"`

	BlockSize       int64  `json:"blockSize"`
	BlocksTotal     uint64 `json:"blocksTotal"`
	BlocksUsed      uint64 `json:blocksUsed`
	BlocksFree      uint64 `json:"blocksFree"`
	BlocksAvailable uint64 `json:"blocksAvailable"`

	FilesTotal uint64 `json:"filesTotal"`
	FilesUsed  uint64 `json:"filesUsed"`
	FilesFree  uint64 `json:"filesFree"`
}

func handleCacheUsage(*http.Request) (*CacheUsageResponse, error) {
	var s unix.Statfs_t
	err := unix.Statfs("/var/cache/fscache", &s)
	if err != nil {
		return nil, err
	}

	bs := uint64(s.Bsize)
	return &CacheUsageResponse{
		BytesTotal:     s.Blocks * bs,
		BytesUsed:      (s.Blocks - s.Bfree) * bs,
		BytesFree:      s.Bfree * bs,
		BytesAvailable: s.Bavail * bs,

		BlockSize:       s.Bsize,
		BlocksTotal:     s.Blocks,
		BlocksUsed:      s.Blocks - s.Bfree,
		BlocksFree:      s.Bfree,
		BlocksAvailable: s.Bavail,

		FilesTotal: s.Files,
		FilesUsed:  s.Files - s.Ffree,
		FilesFree:  s.Ffree,
	}, nil
}
