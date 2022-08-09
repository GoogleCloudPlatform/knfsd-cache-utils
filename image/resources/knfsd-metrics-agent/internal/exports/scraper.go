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

package exports

import (
	"context"
	"time"

	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-metrics-agent/convert"
	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-metrics-agent/internal/exports/internal/metadata"
	"github.com/prometheus/procfs/nfs"
	"go.opentelemetry.io/collector/model/pdata"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

type exportScraper struct {
	mb metadata.MetricsBuilder
}

func newScraper(cfg *Config) (scraperhelper.Scraper, error) {
	s := &exportScraper{
		mb: *metadata.NewMetricsBuilder(cfg.Metrics),
	}
	return scraperhelper.NewScraper(typeStr, s.scrape)
}

func (s *exportScraper) scrape(context.Context) (pdata.Metrics, error) {
	md := pdata.NewMetrics()
	now := pdata.NewTimestampFromTime(time.Now())

	fs, err := nfs.NewDefaultFS()
	if err != nil {
		return md, err
	}

	stats, err := fs.ServerRPCStats()
	if err != nil {
		return md, err
	}

	metrics := md.ResourceMetrics().AppendEmpty().
		InstrumentationLibraryMetrics().AppendEmpty().
		Metrics()

	ops := totalOperations(stats)

	s.mb.RecordNfsExportsTotalOperationsDataPoint(now, convert.Int64(ops))
	s.mb.RecordNfsExportsTotalReadBytesDataPoint(now, convert.Int64(stats.InputOutput.Read))
	s.mb.RecordNfsExportsTotalWriteBytesDataPoint(now, convert.Int64(stats.InputOutput.Write))
	s.mb.Emit(metrics)
	return md, nil
}

func totalOperations(stats *nfs.ServerRPCStats) uint64 {
	var ops uint64

	// Total up all the operations. Not using stats.ServerRPC.RPCCount because
	// NFSv4 can have multiple operations in a single RPC call as NFSv4 only has
	// two RPC calls, NULL and COMPOUND.
	// ignore V2Stats, NFS v2 is obsolete

	ops += stats.V3Stats.Null
	ops += stats.V3Stats.GetAttr
	ops += stats.V3Stats.SetAttr
	ops += stats.V3Stats.Lookup
	ops += stats.V3Stats.Access
	ops += stats.V3Stats.ReadLink
	ops += stats.V3Stats.Read
	ops += stats.V3Stats.Write
	ops += stats.V3Stats.Create
	ops += stats.V3Stats.MkDir
	ops += stats.V3Stats.SymLink
	ops += stats.V3Stats.MkNod
	ops += stats.V3Stats.Remove
	ops += stats.V3Stats.RmDir
	ops += stats.V3Stats.Rename
	ops += stats.V3Stats.Link
	ops += stats.V3Stats.ReadDir
	ops += stats.V3Stats.ReadDirPlus
	ops += stats.V3Stats.FsStat
	ops += stats.V3Stats.FsInfo
	ops += stats.V3Stats.PathConf
	ops += stats.V3Stats.Commit

	ops += stats.ServerV4Stats.Null
	// ignore ServerV4Stats.Compound as it only groups the operations below
	ops += stats.V4Ops.Access
	ops += stats.V4Ops.Close
	ops += stats.V4Ops.Commit
	ops += stats.V4Ops.Create
	ops += stats.V4Ops.DelegPurge
	ops += stats.V4Ops.DelegReturn
	ops += stats.V4Ops.GetAttr
	ops += stats.V4Ops.GetFH
	ops += stats.V4Ops.Link
	ops += stats.V4Ops.Lock
	ops += stats.V4Ops.Lockt
	ops += stats.V4Ops.Locku
	ops += stats.V4Ops.Lookup
	ops += stats.V4Ops.LookupRoot
	ops += stats.V4Ops.Nverify
	ops += stats.V4Ops.Open
	ops += stats.V4Ops.OpenAttr
	ops += stats.V4Ops.OpenConfirm
	ops += stats.V4Ops.OpenDgrd
	ops += stats.V4Ops.PutFH
	ops += stats.V4Ops.PutPubFH
	ops += stats.V4Ops.PutRootFH
	ops += stats.V4Ops.Read
	ops += stats.V4Ops.ReadDir
	ops += stats.V4Ops.ReadLink
	ops += stats.V4Ops.Remove
	ops += stats.V4Ops.Rename
	ops += stats.V4Ops.Renew
	ops += stats.V4Ops.RestoreFH
	ops += stats.V4Ops.SaveFH
	ops += stats.V4Ops.SecInfo
	ops += stats.V4Ops.SetAttr
	ops += stats.V4Ops.Verify
	ops += stats.V4Ops.Write
	ops += stats.V4Ops.RelLockOwner

	return ops
}
