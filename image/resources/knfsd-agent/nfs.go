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

	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-agent/client"
	"github.com/prometheus/procfs"
	"github.com/prometheus/procfs/nfs"
)

func handleNFSClientStats(*http.Request) (*client.NFSClientStats, error) {
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

func handleNFSServerStats(*http.Request) (*client.NFSServerStats, error) {
	fs, err := nfs.NewDefaultFS()
	if err != nil {
		return nil, err
	}
	return readNFSServerStats(fs)
}

func readNFSClientStats(fs nfs.FS, proc procfs.Proc, nfsRoot string) (*client.NFSClientStats, error) {
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

	return &client.NFSClientStats{
		IO: io,
		Network: client.NFSNetwork{
			TotalPackets:   s.Network.NetCount,
			UDPPackets:     s.Network.UDPCount,
			TCPPackets:     s.Network.TCPCount,
			TCPConnections: s.Network.TCPConnect,
		},
		RPC: client.NFSClientRPC{
			Count:           s.ClientRPC.RPCCount,
			AuthRefreshes:   s.ClientRPC.AuthRefreshes,
			Retransmissions: s.ClientRPC.Retransmissions,
		},
		Proc3: client.NFSProc3{
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
		Proc4: client.NFSClientProc4{
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

func readNFSClientIOTotal(proc procfs.Proc, nfsRoot string) (client.NFSIO, error) {
	s, err := proc.MountStats()
	if err != nil {
		return client.NFSIO{}, err
	}

	var io client.NFSIO
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

func readNFSServerStats(fs nfs.FS) (*client.NFSServerStats, error) {
	s, err := fs.ServerRPCStats()
	if err != nil {
		return nil, err
	}

	return &client.NFSServerStats{
		Threads: s.Threads.Threads,
		IO: client.NFSIO{
			Read:  s.InputOutput.Read,
			Write: s.InputOutput.Write,
		},
		Network: client.NFSNetwork{
			TotalPackets:   s.Network.NetCount,
			UDPPackets:     s.Network.UDPCount,
			TCPPackets:     s.Network.TCPCount,
			TCPConnections: s.Network.TCPConnect,
		},
		RPC: client.NFSServerRPC{
			Count:     s.ServerRPC.RPCCount,
			BadTotal:  s.ServerRPC.BadCnt,
			BadFormat: s.ServerRPC.BadFmt,
			BadAuth:   s.ServerRPC.BadAuth,
		},
		Proc3: client.NFSProc3{
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
		Proc4: client.NFSServerProc4{
			Null:     s.ServerV4Stats.Null,
			Compound: s.ServerV4Stats.Compound,
		},
		Proc4Ops: client.NFSProc4Ops{
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
