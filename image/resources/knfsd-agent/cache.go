/*
 Copyright 2022 Google LLC

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package main

import (
	"net/http"

	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-agent/client"
	"golang.org/x/sys/unix"
)

func handleCacheUsage(*http.Request) (*client.CacheUsageResponse, error) {
	var s unix.Statfs_t
	err := unix.Statfs("/var/cache/fscache", &s)
	if err != nil {
		return nil, err
	}

	bs := uint64(s.Bsize)
	return &client.CacheUsageResponse{
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
