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

type NFSClientStats struct {
	IO      NFSIO          `json:"io"`
	Network NFSNetwork     `json:"net"`
	RPC     NFSClientRPC   `json:"rpc"`
	Proc3   NFSProc3       `json:"proc3"`
	Proc4   NFSClientProc4 `json:"proc4"`
}

type NFSServerStats struct {
	Threads  uint64         `json:"threads"`
	IO       NFSIO          `json:"io"`
	Network  NFSNetwork     `json:"net"`
	RPC      NFSServerRPC   `json:"rpc"`
	Proc3    NFSProc3       `json:"proc3"`
	Proc4    NFSServerProc4 `json:"proc4"`
	Proc4Ops NFSProc4Ops    `json:"proc4ops"`
}

type NFSIO struct {
	Read  uint64 `json:"read"`
	Write uint64 `json:"write"`
}

type NFSNetwork struct {
	TotalPackets   uint64 `json:"totalPackets"`
	UDPPackets     uint64 `json:"udpPackets"`
	TCPPackets     uint64 `json:"tcpPackets"`
	TCPConnections uint64 `json:"tcpConnections"`
}

type NFSClientRPC struct {
	Count           uint64 `json:"count"`
	AuthRefreshes   uint64 `json:"authRefreshes"`
	Retransmissions uint64 `json:"retransmissions"`
}

type NFSServerRPC struct {
	Count     uint64 `json:"count"`
	BadTotal  uint64 `json:"badTotal"`
	BadFormat uint64 `json:"badFormat"`
	BadAuth   uint64 `json:"badAuth"`
}

// Use uppercase for JSON field names to match the operations listed in
// mountstats. This is consistent with how the NFS protocol documentation refers
// to operation names.
// This will allow clients to treat proc3/proc4/proc4ops as simple maps of NFS
// procedure name to uint64 counter.
type NFSProc3 struct {
	Null        uint64 `json:"NULL"`
	GetAttr     uint64 `json:"GETATTR"`
	SetAttr     uint64 `json:"SETATTR"`
	Lookup      uint64 `json:"LOOKUP"`
	Access      uint64 `json:"ACCESS"`
	ReadLink    uint64 `json:"READLINK"`
	Read        uint64 `json:"READ"`
	Write       uint64 `json:"WRITE"`
	Create      uint64 `json:"CREATE"`
	MkDir       uint64 `json:"MKDIR"`
	SymLink     uint64 `json:"SYMLINK"`
	MkNod       uint64 `json:"MKNOD"`
	Remove      uint64 `json:"REMOVE"`
	RmDir       uint64 `json:"RMDIR"`
	Rename      uint64 `json:"RENAME"`
	Link        uint64 `json:"LINK"`
	ReadDir     uint64 `json:"READDIR"`
	ReadDirPlus uint64 `json:"READDIRPLUS"`
	FsStat      uint64 `json:"FSSTAT"`
	FsInfo      uint64 `json:"FSINFO"`
	PathConf    uint64 `json:"PATHCONF"`
	Commit      uint64 `json:"COMMIT"`
}

type NFSClientProc4 struct {
	Null               uint64 `json:"NULL"`
	Read               uint64 `json:"READ"`
	Write              uint64 `json:"WRITE"`
	Commit             uint64 `json:"COMMIT"`
	Open               uint64 `json:"OPEN"`
	OpenConfirm        uint64 `json:"OPEN_CONFIRM"`
	OpenNoattr         uint64 `json:"OPEN_NOATTR"`
	OpenDowngrade      uint64 `json:"OPEN_DOWNGRADE"`
	Close              uint64 `json:"CLOSE"`
	Setattr            uint64 `json:"SETATTR"`
	FsInfo             uint64 `json:"FSINFO"`
	Renew              uint64 `json:"RENEW"`
	SetClientID        uint64 `json:"SETCLIENTID"`
	SetClientIDConfirm uint64 `json:"SETCLIENTID_CONFIRM"`
	Lock               uint64 `json:"LOCK"`
	Lockt              uint64 `json:"LOCKT"`
	Locku              uint64 `json:"LOCKU"`
	Access             uint64 `json:"ACCESS"`
	Getattr            uint64 `json:"GETATTR"`
	Lookup             uint64 `json:"LOOKUP"`
	LookupRoot         uint64 `json:"LOOKUP_ROOT"`
	Remove             uint64 `json:"REMOVE"`
	Rename             uint64 `json:"RENAME"`
	Link               uint64 `json:"LINK"`
	Symlink            uint64 `json:"SYMLINK"`
	Create             uint64 `json:"CREATE"`
	Pathconf           uint64 `json:"PATHCONF"`
	StatFs             uint64 `json:"STATFS"`
	ReadLink           uint64 `json:"READLINK"`
	ReadDir            uint64 `json:"READDIR"`
	ServerCaps         uint64 `json:"SERVER_CAPS"`
	DelegReturn        uint64 `json:"DELEGRETURN"`
	GetACL             uint64 `json:"GETACL"`
	SetACL             uint64 `json:"SETACL"`
	FsLocations        uint64 `json:"FS_LOCATIONS"`
	ReleaseLockowner   uint64 `json:"RELEASE_LOCKOWNER"`
	Secinfo            uint64 `json:"SECINFO"`
	FsidPresent        uint64 `json:"FSID_PRESENT"`
	ExchangeID         uint64 `json:"EXCHANGE_ID"`
	CreateSession      uint64 `json:"CREATE_SESSION"`
	DestroySession     uint64 `json:"DESTROY_SESSION"`
	Sequence           uint64 `json:"SEQUENCE"`
	GetLeaseTime       uint64 `json:"GET_LEASE_TIME"`
	ReclaimComplete    uint64 `json:"RECLAIM_COMPLETE"`
	LayoutGet          uint64 `json:"GETDEVICEINFO"`
	GetDeviceInfo      uint64 `json:"LAYOUTGET"`
	LayoutCommit       uint64 `json:"LAYOUTCOMMIT"`
	LayoutReturn       uint64 `json:"LAYOUTRETURN"`
	SecinfoNoName      uint64 `json:"SECINFO_NO_NAME"`
	TestStateID        uint64 `json:"TEST_STATEID"`
	FreeStateID        uint64 `json:"FREE_STATEID"`
	GetDeviceList      uint64 `json:"GETDEVICELIST"`
	BindConnToSession  uint64 `json:"BIND_CONN_TO_SESSION"`
	DestroyClientID    uint64 `json:"DESTROY_CLIENTID"`
	Seek               uint64 `json:"SEEK"`
	Allocate           uint64 `json:"ALLOCATE"`
	DeAllocate         uint64 `json:"DEALLOCATE"`
	LayoutStats        uint64 `json:"LAYOUTSTATS"`
	Clone              uint64 `json:"CLONE"`
}

type NFSServerProc4 struct {
	Null     uint64 `json:"NULL"`
	Compound uint64 `json:"COMPOUND"`
}

type NFSProc4Ops struct {
	Access             uint64 `json:"ACCESS"`
	Close              uint64 `json:"CLOSE"`
	Commit             uint64 `json:"COMMIT"`
	Create             uint64 `json:"CREATE"`
	DelegPurge         uint64 `json:"DELEGPURGE"`
	DelegReturn        uint64 `json:"DELEGRETURN"`
	GetAttr            uint64 `json:"GETATTR"`
	GetFH              uint64 `json:"GETFH"`
	Link               uint64 `json:"LINK"`
	Lock               uint64 `json:"LOCK"`
	LockTest           uint64 `json:"LOCKT"`
	Unlock             uint64 `json:"LOCKU"`
	Lookup             uint64 `json:"LOOKUP"`
	LookupParent       uint64 `json:"LOOKUPP"`
	NVerify            uint64 `json:"NVERIFY"`
	Open               uint64 `json:"OPEN"`
	OpenAttr           uint64 `json:"OPENATTR"`
	OpenConfirm        uint64 `json:"OPEN_CONFIRM"`
	OpenDowngrade      uint64 `json:"OPEN_DOWNGRADE"`
	PutFH              uint64 `json:"PUTFH"`
	PutPubFH           uint64 `json:"PUTPUBFH"`
	PutRootFH          uint64 `json:"PUTROOTFH"`
	Read               uint64 `json:"READ"`
	ReadDir            uint64 `json:"READDIR"`
	ReadLink           uint64 `json:"READLINK"`
	Remove             uint64 `json:"REMOVE"`
	Rename             uint64 `json:"RENAME"`
	Renew              uint64 `json:"RENEW"`
	RestoreFH          uint64 `json:"RESTOREFH"`
	SaveFH             uint64 `json:"SAVEFH"`
	SecInfo            uint64 `json:"SECINFO"`
	SetAttr            uint64 `json:"SETATTR"`
	SetClientID        uint64 `json:"SETCLIENTID"`
	SetClientIDConfirm uint64 `json:"SETCLIENTID_CONFIRM"`
	Verify             uint64 `json:"VERIFY"`
	Write              uint64 `json:"WRITE"`
	ReleaseLockOwner   uint64 `json:"RELEASE_LOCKOWNER"`
}

func (c *KnfsdAgentClient) NFSClientStats() (*NFSClientStats, error) {
	var v *NFSClientStats
	err := c.get("api/v1/nfs/client", &v)
	return v, err
}

func (c *KnfsdAgentClient) NFSServerStats() (*NFSServerStats, error) {
	var v *NFSServerStats
	err := c.get("api/v1/nfs/server", &v)
	return v, err
}
