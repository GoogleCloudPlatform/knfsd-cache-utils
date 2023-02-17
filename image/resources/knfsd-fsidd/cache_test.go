package main

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
)

type any = interface{}

type OpGetFSID struct{ path string }
type OpAllocateFSID struct{ path string }
type OpGetPath struct{ fsid int32 }

type FakeSource struct {
	called []any
}

func (s *FakeSource) GetFSID(ctx context.Context, path string) (int32, error) {
	s.called = append(s.called, OpGetFSID{path})
	switch path {
	case "/foo":
		return 1, nil
	case "/bar":
		return 2, nil
	default:
		return 0, pgx.ErrNoRows
	}
}

func (s *FakeSource) AllocateFSID(ctx context.Context, path string) (int32, error) {
	s.called = append(s.called, OpAllocateFSID{path})
	switch path {
	case "/foo":
		return 1, nil
	case "/bar":
		return 2, nil
	default:
		return 0, pgx.ErrNoRows
	}
}

func (s *FakeSource) GetPath(ctx context.Context, fsid int32) (string, error) {
	s.called = append(s.called, OpGetPath{fsid})
	switch fsid {
	case 1:
		return "/foo", nil
	case 2:
		return "/bar", nil
	default:
		return "", pgx.ErrNoRows
	}
}

type CacheTester struct {
	t      *testing.T
	source *FakeSource
	cache  *FSIDCache
}

func (t *CacheTester) GetFSID(path string) *CacheResult {
	fsid, err := t.cache.GetFSID(context.Background(), path)
	return t.result(fsid, err)
}

func (t *CacheTester) AllocateFSID(path string) *CacheResult {
	fsid, err := t.cache.AllocateFSID(context.Background(), path)
	return t.result(fsid, err)
}

func (t *CacheTester) GetPath(fsid int32) *CacheResult {
	path, err := t.cache.GetPath(context.Background(), fsid)
	return t.result(path, err)
}

func (t *CacheTester) result(value interface{}, err error) *CacheResult {
	called := t.source.called
	t.source.called = nil
	return &CacheResult{
		t:      t.t,
		value:  value,
		err:    err,
		called: called,
	}
}

type CacheResult struct {
	t      *testing.T
	value  any
	err    error
	called []any
}

func (r *CacheResult) Ok(expected any) *CacheResult {
	if v, ok := expected.(int); ok {
		// makes test syntax nicer, allows Ok(1) to match an fsid, otherwise
		// the syntax would be Ok(int32(1))
		expected = int32(v)
	}

	assert.NoError(r.t, r.err)
	assert.Equal(r.t, expected, r.value)
	return r
}

func (r *CacheResult) Err() *CacheResult {
	assert.Error(r.t, r.err)
	return r
}

func (r *CacheResult) WasCalled(expected ...any) *CacheResult {
	assert.Equal(r.t, expected, r.called)
	return r
}

func (r *CacheResult) NotCalled() *CacheResult {
	assert.Empty(r.t, r.called)
	return r
}

func TestFSIDCache(t *testing.T) {
	newTest := func(t *testing.T) CacheTester {
		source := FakeSource{}
		cache := FSIDCache{source: &source}
		return CacheTester{t: t, source: &source, cache: &cache}
	}

	// Check the responses are cached. After the first call to resolve an fsid
	// or path both the fsid and path should be cached. Thus only the first
	// call to any of these methods should go to the source.

	t.Run("GetFSID", func(t *testing.T) {
		test := newTest(t)
		test.GetFSID("/foo").Ok(1).WasCalled(OpGetFSID{"/foo"})

		test.GetFSID("/foo").Ok(1).NotCalled()
		test.AllocateFSID("/foo").Ok(1).NotCalled()
		test.GetPath(1).Ok("/foo").NotCalled()

		test.GetFSID("/bar").Ok(2).WasCalled(OpGetFSID{"/bar"})
	})

	t.Run("AllocatedFSID", func(t *testing.T) {
		test := newTest(t)
		test.AllocateFSID("/foo").Ok(1).WasCalled(OpAllocateFSID{"/foo"})

		test.GetFSID("/foo").Ok(1).NotCalled()
		test.AllocateFSID("/foo").Ok(1).NotCalled()
		test.GetPath(1).Ok("/foo").NotCalled()

		test.AllocateFSID("/bar").Ok(2).WasCalled(OpAllocateFSID{"/bar"})
	})

	t.Run("GetPath", func(t *testing.T) {
		test := newTest(t)
		test.GetPath(1).Ok("/foo").WasCalled(OpGetPath{1})

		test.GetFSID("/foo").Ok(1).NotCalled()
		test.AllocateFSID("/foo").Ok(1).NotCalled()
		test.GetPath(1).Ok("/foo").NotCalled()

		test.GetPath(2).Ok("/bar").WasCalled(OpGetPath{2})
	})

	t.Run("errors not cached", func(t *testing.T) {
		test := newTest(t)

		// Error responses should not be cached, so calling the same method
		// twice should go to the source both times.

		test.GetFSID("/unknown").Err().WasCalled(OpGetFSID{"/unknown"})
		test.GetFSID("/unknown").Err().WasCalled(OpGetFSID{"/unknown"})

		test.AllocateFSID("/unknown").Err().WasCalled(OpAllocateFSID{"/unknown"})
		test.AllocateFSID("/unknown").Err().WasCalled(OpAllocateFSID{"/unknown"})

		test.GetPath(0).Err().WasCalled(OpGetPath{0})
		test.GetPath(0).Err().WasCalled(OpGetPath{0})
	})
}
