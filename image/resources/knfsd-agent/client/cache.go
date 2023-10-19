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

package client

type CacheUsageResponse struct {
	BytesTotal     uint64 `json:"bytesTotal"`
	BytesUsed      uint64 `json:"bytesUsed"`
	BytesFree      uint64 `json:"bytesFree"`
	BytesAvailable uint64 `json:"bytesAvailable"`

	BlockSize       int64  `json:"blockSize"`
	BlocksTotal     uint64 `json:"blocksTotal"`
	BlocksUsed      uint64 `json:"blocksUsed"`
	BlocksFree      uint64 `json:"blocksFree"`
	BlocksAvailable uint64 `json:"blocksAvailable"`

	FilesTotal uint64 `json:"filesTotal"`
	FilesUsed  uint64 `json:"filesUsed"`
	FilesFree  uint64 `json:"filesFree"`
}

func (c *KnfsdAgentClient) CacheUsage() (*CacheUsageResponse, error) {
	var v *CacheUsageResponse
	err := c.get("api/v1/cache/usage", &v)
	return v, err
}
