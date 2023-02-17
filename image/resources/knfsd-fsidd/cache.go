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
	"context"
	"sync"
)

// FSIDCache provides a simple read cache improve read performance by avoiding
// querying the remote database for fsids that have already been resolved by
// previous requests.
// If there are concurrent requests to get or allocate an FSID, let those
// requests race each other. Concurrency and consistency will be handled by the
// database.
type FSIDCache struct {
	source FSIDProvider
	fsids  sync.Map // path => fsid
	paths  sync.Map // fsid => path
}

func (c *FSIDCache) GetFSID(ctx context.Context, path string) (int32, error) {
	if fsid, ok := c.fsids.Load(path); ok {
		return fsid.(int32), nil
	}
	fsid, err := c.source.GetFSID(ctx, path)
	if err == nil {
		c.store(fsid, path)
	}
	return fsid, err
}

func (c *FSIDCache) AllocateFSID(ctx context.Context, path string) (int32, error) {
	if fsid, ok := c.fsids.Load(path); ok {
		return fsid.(int32), nil
	}
	fsid, err := c.source.AllocateFSID(ctx, path)
	if err == nil {
		c.store(fsid, path)
	}
	return fsid, err
}

func (c *FSIDCache) GetPath(ctx context.Context, fsid int32) (string, error) {
	if path, ok := c.paths.Load(fsid); ok {
		return path.(string), nil
	}
	path, err := c.source.GetPath(ctx, fsid)
	if err == nil {
		c.store(fsid, path)
	}
	return path, err
}

func (c *FSIDCache) store(fsid int32, path string) {
	// Concurrent requests will all result in the same fsid, path pair so just
	// blindly store the results.
	// There may be some initial contention on the keys but that will quickly
	// resolve itself, after which the fsid/path combination will be readonly.
	c.fsids.Store(path, fsid)
	c.paths.Store(fsid, path)
}
