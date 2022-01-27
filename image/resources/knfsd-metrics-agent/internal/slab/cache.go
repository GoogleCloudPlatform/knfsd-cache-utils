package slab

import (
	"context"
	"time"

	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-metrics-agent/internal/collectd"
	"github.com/prometheus/procfs"
)

type cacheCounter struct {
	name          string
	activeObjects collectd.Gauge
	objectSize    collectd.Gauge
}

var (
	cacheCounters = []cacheCounter{
		{
			name:          "nfs_inode_cache",
			activeObjects: collectd.NewGauge("nfs_inode_cache_active_objects", "usage", nil),
			objectSize:    collectd.NewGauge("nfs_inode_cache_objsize", "usage", nil),
		},
		{
			name:          "dentry",
			activeObjects: collectd.NewGauge("dentry_cache_active_objects", "usage", nil),
			objectSize:    collectd.NewGauge("dentry_cache_objsize", "usage", nil),
		},
	}
)

func Report(ctx context.Context, fs procfs.FS) error {
	info, err := fs.SlabInfo()
	if err != nil {
		return err
	}

	now := time.Now()
	for _, c := range cacheCounters {
		s := find(info.Slabs, c.name)
		if s != nil {
			reportCacheEntry(ctx, now, c, s)
		}
	}
	return nil
}

func reportCacheEntry(
	ctx context.Context,
	now time.Time,
	c cacheCounter,
	s *procfs.Slab,
) {
	c.activeObjects.Write(ctx, now, float64(s.ObjActive))
	c.objectSize.Write(ctx, now, float64(s.ObjSize))
}

func find(slabs []*procfs.Slab, name string) *procfs.Slab {
	for _, s := range slabs {
		if s.Name == name {
			return s
		}
	}
	return nil
}
