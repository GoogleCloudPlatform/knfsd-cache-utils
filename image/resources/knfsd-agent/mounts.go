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
	"encoding"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/prometheus/procfs"
)

var (
	_ encoding.TextMarshaler   = (*Duration)(nil)
	_ encoding.TextUnmarshaler = (*Duration)(nil)
)

type Duration time.Duration

func (d Duration) MarshalText() ([]byte, error) {
	return []byte((time.Duration)(d).String()), nil
}

func (d *Duration) UnmarshalText(text []byte) error {
	v, err := time.ParseDuration(string(text))
	*d = Duration(v)
	return err
}

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

func handleMounts(*http.Request) (*MountResponse, error) {
	nfsRoot, err := getNFSRootDir()
	if err != nil {
		return nil, err
	}

	self, err := procfs.Self()
	if err != nil {
		return nil, err
	}

	res, err := readMounts(self, nfsRoot)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func readMounts(proc procfs.Proc, nfsRoot string) (*MountResponse, error) {
	info, err := proc.MountInfo()
	if err != nil {
		return nil, err
	}

	var mounts []Mount
	for _, e := range info {
		if !isNFS(e.FSType) {
			continue
		}
		if !strings.HasPrefix(e.MountPoint, nfsRoot) {
			continue
		}

		m := Mount{
			Device:  e.Source,
			Mount:   e.MountPoint,
			Export:  e.MountPoint[len(nfsRoot)-1:],
			Options: combineMountOptions(e.Options, e.SuperOptions),
		}
		mounts = append(mounts, m)
	}

	return &MountResponse{mounts}, nil
}

func handleMountStats(*http.Request) (*MountStatsResponse, error) {
	nfsRoot, err := getNFSRootDir()
	if err != nil {
		return nil, err
	}

	self, err := procfs.Self()
	if err != nil {
		return nil, err
	}

	res, err := readMountStats(self, nfsRoot)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func readMountStats(proc procfs.Proc, nfsRoot string) (*MountStatsResponse, error) {
	info, err := proc.MountInfo()
	if err != nil {
		return nil, err
	}

	stats, err := proc.MountStats()
	if err != nil {
		return nil, err
	}

	type InfoKey struct {
		Device string
		Mount  string
	}

	// NFS stats only includes super options and is missing some of the
	// per-mount options such as noatime. So fetch all the of the options from
	// /proc/self/mountinfo.
	// This avoids a client needing to query both /mounts and /mountstats and
	// having to merge the results locally to get the complete information.
	options := make(map[InfoKey]map[string]string)
	for _, m := range info {
		if !isNFS(m.FSType) {
			continue
		}
		if !strings.HasPrefix(m.MountPoint, nfsRoot) {
			continue
		}
		key := InfoKey{
			Device: m.Source,
			Mount:  m.MountPoint,
		}
		options[key] = combineMountOptions(m.Options, m.SuperOptions)
	}

	var mounts []MountStats
	for _, e := range stats {
		if !isNFS(e.Type) {
			continue
		}
		if !strings.HasPrefix(e.Mount, nfsRoot) {
			continue
		}

		m := MountStats{
			Device: e.Device,
			Mount:  e.Mount,
			Export: e.Mount[len(nfsRoot)-1:],
			// lookup the options from mountinfo
			Options: options[InfoKey{e.Device, e.Mount}],
		}

		if s, ok := e.Stats.(*procfs.MountStatsNFS); ok {
			// combine the options from the stats with options from mountinfo
			m.Options = combineMountOptions(m.Options, s.Opts)

			ops := make([]NFSOperationStats, len(s.Operations))
			for i, o := range s.Operations {
				retries := uint64(0)
				if o.Transmissions > o.Requests {
					retries = o.Transmissions - o.Requests
				}

				ops[i] = NFSOperationStats{
					Operation:             o.Operation,
					Requests:              o.Requests,
					Transmissions:         o.Transmissions,
					Retries:               retries,
					MajorTimeouts:         o.MajorTimeouts,
					BytesSent:             o.BytesSent,
					BytesReceived:         o.BytesReceived,
					QueueMilliseconds:     o.CumulativeQueueMilliseconds,
					RTTMilliseconds:       o.CumulativeTotalResponseMilliseconds,
					ExecutionMilliseconds: o.CumulativeTotalRequestMilliseconds,
					Errors:                o.Errors,
				}
			}

			m.Stats = NFSMountStats{
				Age: Duration(s.Age),
				Bytes: NFSByteStats{
					NormalRead:  s.Bytes.Read,
					NormalWrite: s.Bytes.Write,
					DirectRead:  s.Bytes.DirectRead,
					DirectWrite: s.Bytes.DirectWrite,
					ServerRead:  s.Bytes.ReadTotal,
					ServerWrite: s.Bytes.WriteTotal,
					ReadPages:   s.Bytes.ReadPages,
					WritePages:  s.Bytes.WritePages,
				},
				Events: NFSEventStats{
					InodeRevalidate:     s.Events.InodeRevalidate,
					DnodeRevalidate:     s.Events.DnodeRevalidate,
					DataInvalidate:      s.Events.DataInvalidate,
					AttributeInvalidate: s.Events.AttributeInvalidate,

					VFSOpen:        s.Events.VFSOpen,
					VFSLookup:      s.Events.VFSLookup,
					VFSAccess:      s.Events.VFSAccess,
					VFSUpdatePage:  s.Events.VFSUpdatePage,
					VFSReadPage:    s.Events.VFSReadPage,
					VFSReadPages:   s.Events.VFSReadPages,
					VFSWritePage:   s.Events.VFSWritePage,
					VFSWritePages:  s.Events.VFSWritePages,
					VFSGetdents:    s.Events.VFSGetdents,
					VFSSetattr:     s.Events.VFSSetattr,
					VFSFlush:       s.Events.VFSFlush,
					VFSFsync:       s.Events.VFSFsync,
					VFSLock:        s.Events.VFSLock,
					VFSFileRelease: s.Events.VFSFileRelease,

					// CongestionWait is not used by the kernel
					Truncation:     s.Events.Truncation,
					WriteExtension: s.Events.WriteExtension,
					SillyRename:    s.Events.SillyRename,
					ShortRead:      s.Events.ShortRead,
					ShortWrite:     s.Events.ShortWrite,
					Delay:          s.Events.JukeboxDelay,
					PNFSRead:       s.Events.PNFSRead,
					PNFSWrite:      s.Events.PNFSWrite,
				},
				Operations: ops,
			}
		}
		mounts = append(mounts, m)
	}

	sort.Slice(mounts, func(i, j int) bool {
		a := mounts[i].Mount
		b := mounts[j].Mount
		return a < b
	})

	return &MountStatsResponse{mounts}, nil
}

func combineMountOptions(opts ...map[string]string) map[string]string {
	combined := make(map[string]string)
	for _, o := range opts {
		for k, v := range o {
			combined[k] = v
		}
	}
	return combined
}
