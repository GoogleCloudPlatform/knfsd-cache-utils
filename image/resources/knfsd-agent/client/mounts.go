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

type MountResponse struct {
	Mounts []Mount `json:"mounts"`
}

type Mount struct {
	Device  string            `json:"device"`
	Mount   string            `json:"mount"`
	Export  string            `json:"export"`
	Options map[string]string `json:"options"`
}

type MountStatsResponse struct {
	Mounts []MountStats `json:"mounts"`
}

type MountStats struct {
	Device  string            `json:"device"`
	Mount   string            `json:"mount"`
	Export  string            `json:"export"`
	Options map[string]string `json:"options"`
	Stats   NFSMountStats     `json:"stats"`
}

type NFSMountStats struct {
	Age        Duration            `json:"age"`
	Bytes      NFSByteStats        `json:"bytes"`
	Events     NFSEventStats       `json:"events"`
	Operations []NFSOperationStats `json:"operations"`
}

type NFSByteStats struct {
	NormalRead  uint64 `json:"normalRead"`
	NormalWrite uint64 `json:"normalWrite"`
	DirectRead  uint64 `json:"directRead"`
	DirectWrite uint64 `json:"directWrite"`
	ServerRead  uint64 `json:"serverRead"`
	ServerWrite uint64 `json:"serverWrite"`
	ReadPages   uint64 `json:"readPages"`
	WritePages  uint64 `json:"writePages"`
}

type NFSEventStats struct {
	InodeRevalidate     uint64 `json:"inodeRevalidate"`
	DnodeRevalidate     uint64 `json:"dnodeRevalidate"`
	DataInvalidate      uint64 `json:"dataInvalidate"`
	AttributeInvalidate uint64 `json:"attributeInvalidate"`
	VFSOpen             uint64 `json:"vfsOpen"`
	VFSLookup           uint64 `json:"vfsLookup"`
	VFSAccess           uint64 `json:"vfsAccess"`
	VFSUpdatePage       uint64 `json:"vfsUpdatePage"`
	VFSReadPage         uint64 `json:"vfsReadPage"`
	VFSReadPages        uint64 `json:"vfsReadPages"`
	VFSWritePage        uint64 `json:"vfsWritePage"`
	VFSWritePages       uint64 `json:"vfsWritePages"`
	VFSGetdents         uint64 `json:"vfsGetdents"`
	VFSSetattr          uint64 `json:"vfsSetattr"`
	VFSFlush            uint64 `json:"vfsFlush"`
	VFSFsync            uint64 `json:"vfsFsync"`
	VFSLock             uint64 `json:"vfsLock"`
	VFSFileRelease      uint64 `json:"vfsFileRelease"`
	Truncation          uint64 `json:"truncation"`
	WriteExtension      uint64 `json:"writeExtension"`
	SillyRename         uint64 `json:"sillyRename"`
	ShortRead           uint64 `json:"shortRead"`
	ShortWrite          uint64 `json:"shortWrite"`
	Delay               uint64 `json:"delay"`
	PNFSRead            uint64 `json:"pnfsRead"`
	PNFSWrite           uint64 `json:"pnfsWrite"`
}

type NFSOperationStats struct {
	Operation             string `json:"operation"`
	Requests              uint64 `json:"requests"`
	Transmissions         uint64 `json:"transmissions"`
	Retries               uint64 `json:"retries"`
	MajorTimeouts         uint64 `json:"majorTimeouts"`
	BytesSent             uint64 `json:"bytesSent"`
	BytesReceived         uint64 `json:"bytesReceived"`
	QueueMilliseconds     uint64 `json:"queueMilliseconds"`
	RTTMilliseconds       uint64 `json:"rttMilliseconds"`
	ExecutionMilliseconds uint64 `json:"executionMilliseconds"`
	Errors                uint64 `json:"errors"`
}

func (c *KnfsdAgentClient) GetMounts() (*MountResponse, error) {
	var v *MountResponse
	err := c.get("api/v1/mounts", &v)
	return v, err
}

func (c *KnfsdAgentClient) GetMountStats() (*MountStatsResponse, error) {
	var v *MountStatsResponse
	err := c.get("api/v1/mountStats", &v)
	return v, err
}
