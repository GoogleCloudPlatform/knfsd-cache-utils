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
	"sort"
	"strings"

	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-agent/client"
	"github.com/prometheus/procfs"
)

func handleMounts(*http.Request) (*client.MountResponse, error) {
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

func readMounts(proc procfs.Proc, nfsRoot string) (*client.MountResponse, error) {
	info, err := proc.MountInfo()
	if err != nil {
		return nil, err
	}

	var mounts []client.Mount
	for _, e := range info {
		if !isNFS(e.FSType) {
			continue
		}
		if !strings.HasPrefix(e.MountPoint, nfsRoot) {
			continue
		}

		m := client.Mount{
			Device:  e.Source,
			Mount:   e.MountPoint,
			Export:  e.MountPoint[len(nfsRoot)-1:],
			Options: combineMountOptions(e.Options, e.SuperOptions),
		}
		mounts = append(mounts, m)
	}

	return &client.MountResponse{Mounts: mounts}, nil
}

func handleMountStats(*http.Request) (*client.MountStatsResponse, error) {
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

func readMountStats(proc procfs.Proc, nfsRoot string) (*client.MountStatsResponse, error) {
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

	var mounts []client.MountStats
	for _, e := range stats {
		if !isNFS(e.Type) {
			continue
		}
		if !strings.HasPrefix(e.Mount, nfsRoot) {
			continue
		}

		m := client.MountStats{
			Device: e.Device,
			Mount:  e.Mount,
			Export: e.Mount[len(nfsRoot)-1:],
			// lookup the options from mountinfo
			Options: options[InfoKey{e.Device, e.Mount}],
		}

		if s, ok := e.Stats.(*procfs.MountStatsNFS); ok {
			// combine the options from the stats with options from mountinfo
			m.Options = combineMountOptions(m.Options, s.Opts)

			ops := make([]client.NFSOperationStats, len(s.Operations))
			for i, o := range s.Operations {
				retries := uint64(0)
				if o.Transmissions > o.Requests {
					retries = o.Transmissions - o.Requests
				}

				ops[i] = client.NFSOperationStats{
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

			m.Stats = client.NFSMountStats{
				Age: client.Duration(s.Age),
				Bytes: client.NFSByteStats{
					NormalRead:  s.Bytes.Read,
					NormalWrite: s.Bytes.Write,
					DirectRead:  s.Bytes.DirectRead,
					DirectWrite: s.Bytes.DirectWrite,
					ServerRead:  s.Bytes.ReadTotal,
					ServerWrite: s.Bytes.WriteTotal,
					ReadPages:   s.Bytes.ReadPages,
					WritePages:  s.Bytes.WritePages,
				},
				Events: client.NFSEventStats{
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

	return &client.MountStatsResponse{Mounts: mounts}, nil
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
