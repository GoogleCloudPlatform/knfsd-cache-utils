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
	"strings"

	"github.com/prometheus/procfs"
	"github.com/prometheus/procfs/nfs"
)

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

func handleNFSClientStats(*http.Request) (*NFSClientStats, error) {
	fs, err := nfs.NewDefaultFS()
	if err != nil {
		return nil, err
	}

	self, err := procfs.Self()
	if err != nil {
		return nil, err
	}

	nfsRoot, err := getNFSRootDir()
	if err != nil {
		return nil, err
	}

	return readNFSClientStats(fs, self, nfsRoot)
}

func handleNFSServerStats(*http.Request) (*NFSServerStats, error) {
	fs, err := nfs.NewDefaultFS()
	if err != nil {
		return nil, err
	}
	return readNFSServerStats(fs)
}

func readNFSClientStats(fs nfs.FS, proc procfs.Proc, nfsRoot string) (*NFSClientStats, error) {
	s, err := fs.ClientRPCStats()
	if err != nil {
		return nil, err
	}

	// IO stats not are included in /proc/net/rpc/nfs (unlike nfsd server
	// stats). Instead we'll need to read these from /proc/self/mountstats
	// and aggregate the total read/write byte counters.
	io, err := readNFSClientIOTotal(proc, nfsRoot)
	if err != nil {
		return nil, err
	}

	return &NFSClientStats{
		IO: io,
		Network: NFSNetwork{
			TotalPackets:   s.Network.NetCount,
			UDPPackets:     s.Network.UDPCount,
			TCPPackets:     s.Network.TCPCount,
			TCPConnections: s.Network.TCPConnect,
		},
		RPC: NFSClientRPC{
			Count:           s.ClientRPC.RPCCount,
			AuthRefreshes:   s.ClientRPC.AuthRefreshes,
			Retransmissions: s.ClientRPC.Retransmissions,
		},
		Proc3: NFSProc3{
			Null:        s.V3Stats.Null,
			GetAttr:     s.V3Stats.GetAttr,
			SetAttr:     s.V3Stats.SetAttr,
			Lookup:      s.V3Stats.Lookup,
			Access:      s.V3Stats.Access,
			ReadLink:    s.V3Stats.ReadLink,
			Read:        s.V3Stats.Read,
			Write:       s.V3Stats.Write,
			Create:      s.V3Stats.Create,
			MkDir:       s.V3Stats.MkDir,
			SymLink:     s.V3Stats.SymLink,
			MkNod:       s.V3Stats.MkNod,
			Remove:      s.V3Stats.Remove,
			RmDir:       s.V3Stats.RmDir,
			Rename:      s.V3Stats.Rename,
			Link:        s.V3Stats.Link,
			ReadDir:     s.V3Stats.ReadDir,
			ReadDirPlus: s.V3Stats.ReadDirPlus,
			FsStat:      s.V3Stats.FsStat,
			FsInfo:      s.V3Stats.FsInfo,
			PathConf:    s.V3Stats.PathConf,
			Commit:      s.V3Stats.Commit,
		},
		Proc4: NFSClientProc4{
			Null:               s.ClientV4Stats.Null,
			Read:               s.ClientV4Stats.Read,
			Write:              s.ClientV4Stats.Write,
			Commit:             s.ClientV4Stats.Commit,
			Open:               s.ClientV4Stats.Open,
			OpenConfirm:        s.ClientV4Stats.OpenConfirm,
			OpenNoattr:         s.ClientV4Stats.OpenNoattr,
			OpenDowngrade:      s.ClientV4Stats.OpenDowngrade,
			Close:              s.ClientV4Stats.Close,
			Setattr:            s.ClientV4Stats.Setattr,
			FsInfo:             s.ClientV4Stats.FsInfo,
			Renew:              s.ClientV4Stats.Renew,
			SetClientID:        s.ClientV4Stats.SetClientID,
			SetClientIDConfirm: s.ClientV4Stats.SetClientIDConfirm,
			Lock:               s.ClientV4Stats.Lock,
			Lockt:              s.ClientV4Stats.Lockt,
			Locku:              s.ClientV4Stats.Locku,
			Access:             s.ClientV4Stats.Access,
			Getattr:            s.ClientV4Stats.Getattr,
			Lookup:             s.ClientV4Stats.Lookup,
			LookupRoot:         s.ClientV4Stats.LookupRoot,
			Remove:             s.ClientV4Stats.Remove,
			Rename:             s.ClientV4Stats.Rename,
			Link:               s.ClientV4Stats.Link,
			Symlink:            s.ClientV4Stats.Symlink,
			Create:             s.ClientV4Stats.Create,
			Pathconf:           s.ClientV4Stats.Pathconf,
			StatFs:             s.ClientV4Stats.StatFs,
			ReadLink:           s.ClientV4Stats.ReadLink,
			ReadDir:            s.ClientV4Stats.ReadDir,
			ServerCaps:         s.ClientV4Stats.ServerCaps,
			DelegReturn:        s.ClientV4Stats.DelegReturn,
			GetACL:             s.ClientV4Stats.GetACL,
			SetACL:             s.ClientV4Stats.SetACL,
			FsLocations:        s.ClientV4Stats.FsLocations,
			ReleaseLockowner:   s.ClientV4Stats.ReleaseLockowner,
			Secinfo:            s.ClientV4Stats.Secinfo,
			FsidPresent:        s.ClientV4Stats.FsidPresent,
			ExchangeID:         s.ClientV4Stats.ExchangeID,
			CreateSession:      s.ClientV4Stats.CreateSession,
			DestroySession:     s.ClientV4Stats.DestroySession,
			Sequence:           s.ClientV4Stats.Sequence,
			GetLeaseTime:       s.ClientV4Stats.GetLeaseTime,
			ReclaimComplete:    s.ClientV4Stats.ReclaimComplete,
			LayoutGet:          s.ClientV4Stats.LayoutGet,
			GetDeviceInfo:      s.ClientV4Stats.GetDeviceInfo,
			LayoutCommit:       s.ClientV4Stats.LayoutCommit,
			LayoutReturn:       s.ClientV4Stats.LayoutReturn,
			SecinfoNoName:      s.ClientV4Stats.SecinfoNoName,
			TestStateID:        s.ClientV4Stats.TestStateID,
			FreeStateID:        s.ClientV4Stats.FreeStateID,
			GetDeviceList:      s.ClientV4Stats.GetDeviceList,
			BindConnToSession:  s.ClientV4Stats.BindConnToSession,
			DestroyClientID:    s.ClientV4Stats.DestroyClientID,
			Seek:               s.ClientV4Stats.Seek,
			Allocate:           s.ClientV4Stats.Allocate,
			DeAllocate:         s.ClientV4Stats.DeAllocate,
			LayoutStats:        s.ClientV4Stats.LayoutStats,
			Clone:              s.ClientV4Stats.Clone,
		},
	}, nil
}

func readNFSClientIOTotal(proc procfs.Proc, nfsRoot string) (NFSIO, error) {
	s, err := proc.MountStats()
	if err != nil {
		return NFSIO{}, err
	}

	var io NFSIO
	for _, m := range s {
		if !isNFS(m.Type) {
			continue
		}
		if !strings.HasPrefix(m.Mount, nfsRoot) {
			continue
		}
		if s, ok := m.Stats.(*procfs.MountStatsNFS); ok {
			io.Read += s.Bytes.ReadTotal
			io.Write += s.Bytes.WriteTotal
		}
	}
	return io, nil
}

func readNFSServerStats(fs nfs.FS) (*NFSServerStats, error) {
	s, err := fs.ServerRPCStats()
	if err != nil {
		return nil, err
	}

	return &NFSServerStats{
		Threads: s.Threads.Threads,
		IO: NFSIO{
			Read:  s.InputOutput.Read,
			Write: s.InputOutput.Write,
		},
		Network: NFSNetwork{
			TotalPackets:   s.Network.NetCount,
			UDPPackets:     s.Network.UDPCount,
			TCPPackets:     s.Network.TCPCount,
			TCPConnections: s.Network.TCPConnect,
		},
		RPC: NFSServerRPC{
			Count:     s.ServerRPC.RPCCount,
			BadTotal:  s.ServerRPC.BadCnt,
			BadFormat: s.ServerRPC.BadFmt,
			BadAuth:   s.ServerRPC.BadAuth,
		},
		Proc3: NFSProc3{
			Null:        s.V3Stats.Null,
			GetAttr:     s.V3Stats.GetAttr,
			SetAttr:     s.V3Stats.SetAttr,
			Lookup:      s.V3Stats.Lookup,
			Access:      s.V3Stats.Access,
			ReadLink:    s.V3Stats.ReadLink,
			Read:        s.V3Stats.Read,
			Write:       s.V3Stats.Write,
			Create:      s.V3Stats.Create,
			MkDir:       s.V3Stats.MkDir,
			SymLink:     s.V3Stats.SymLink,
			MkNod:       s.V3Stats.MkNod,
			Remove:      s.V3Stats.Remove,
			RmDir:       s.V3Stats.RmDir,
			Rename:      s.V3Stats.Rename,
			Link:        s.V3Stats.Link,
			ReadDir:     s.V3Stats.ReadDir,
			ReadDirPlus: s.V3Stats.ReadDirPlus,
			FsStat:      s.V3Stats.FsStat,
			FsInfo:      s.V3Stats.FsInfo,
			PathConf:    s.V3Stats.PathConf,
			Commit:      s.V3Stats.Commit,
		},
		Proc4: NFSServerProc4{
			Null:     s.ServerV4Stats.Null,
			Compound: s.ServerV4Stats.Compound,
		},
		Proc4Ops: NFSProc4Ops{
			Access:             s.V4Ops.Access,
			Close:              s.V4Ops.Close,
			Commit:             s.V4Ops.Commit,
			Create:             s.V4Ops.Create,
			DelegPurge:         s.V4Ops.DelegPurge,
			DelegReturn:        s.V4Ops.DelegReturn,
			GetAttr:            s.V4Ops.GetAttr,
			GetFH:              s.V4Ops.GetFH,
			Link:               s.V4Ops.Link,
			Lock:               s.V4Ops.Lock,
			LockTest:           s.V4Ops.Lockt,
			Unlock:             s.V4Ops.Locku,
			Lookup:             s.V4Ops.Lookup,
			LookupParent:       s.V4Ops.LookupRoot,
			NVerify:            s.V4Ops.Nverify,
			Open:               s.V4Ops.Open,
			OpenAttr:           s.V4Ops.OpenAttr,
			OpenConfirm:        s.V4Ops.OpenConfirm,
			OpenDowngrade:      s.V4Ops.OpenDgrd,
			PutFH:              s.V4Ops.PutFH,
			PutPubFH:           s.V4Ops.PutPubFH,
			PutRootFH:          s.V4Ops.PutRootFH,
			Read:               s.V4Ops.Read,
			ReadDir:            s.V4Ops.ReadDir,
			ReadLink:           s.V4Ops.ReadLink,
			Remove:             s.V4Ops.Remove,
			Rename:             s.V4Ops.Rename,
			Renew:              s.V4Ops.Renew,
			RestoreFH:          s.V4Ops.RestoreFH,
			SaveFH:             s.V4Ops.SaveFH,
			SecInfo:            s.V4Ops.SecInfo,
			SetAttr:            s.V4Ops.SetAttr,
			SetClientID:        s.V4Ops.SetClientID,
			SetClientIDConfirm: s.V4Ops.SetClientIDConfirm,
			Verify:             s.V4Ops.Verify,
			Write:              s.V4Ops.Write,
			ReleaseLockOwner:   s.V4Ops.RelLockOwner,
		},
	}, nil
}
