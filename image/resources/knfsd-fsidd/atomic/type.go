package atomic

import "sync/atomic"

type noCopy struct{}

// Lock is a no-op used by -copylocks checker from `go vet`.
func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

type Bool struct {
	_ noCopy
	v uint32
}

func (x *Bool) Load() bool { return atomic.LoadUint32(&x.v) != 0 }

func (x *Bool) Store(b bool) { atomic.StoreUint32(&x.v, b32(b)) }

func b32(b bool) uint32 {
	if b {
		return 1
	} else {
		return 0
	}
}

type Uint64 struct {
	_ noCopy
	v uint64
}

func (x *Uint64) Add(delta uint64) uint64 {
	return atomic.AddUint64(&x.v, delta)
}
