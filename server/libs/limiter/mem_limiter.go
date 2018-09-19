package limiter

import (
	"sync"
	"time"
	"errors"
)

const accessOk = 0
const accessNoToken = 1
const accessShouldReset = 2

var errLimiterIsFull = errors.New("Limiter reachs its capacity")

func now() int64 {
	return time.Now().Unix()
}

func after(duration int64) int64 {
	return now() + duration
}

type Bucket struct {
	ResetAt int64
	Tokens  int
}

func NewBucket(resetCyc int64, tokens int) *Bucket {
	if resetCyc <= 0 || tokens < 0 {
		return nil
	}

	return &Bucket{
		ResetAt: after(resetCyc),
		Tokens:  tokens,
	}
}

func (b *Bucket) Access() int {
	now := now()

	if b.ResetAt > now && b.Tokens == 0 {
		return accessNoToken
	} else if b.ResetAt > now && b.Tokens > 0 {
		return accessOk
	}
	return accessShouldReset
}

func (b *Bucket) Reset(resetCyc int64, resetTokens int) {
	b.ResetAt = after(resetCyc)
	b.Tokens = resetTokens - 1
}

func (b *Bucket) Minus() {
	b.Tokens--
}

type MemLimiter struct {
	buckets   map[string]*Bucket
	cap       int
	mux       sync.RWMutex
	rstCyc    int64
	rstTokens int
}

func New(cap int, rstCyc int64, rstTokens int) *MemLimiter {
	if cap <= 0 || rstCyc <= 0 || rstTokens < 0 {
		return nil
	}

	return &MemLimiter{
		cap:       cap,
		buckets:   make(map[string]*Bucket, 0),
		rstCyc:    rstCyc,
		rstTokens: rstTokens,
	}
}

func (l *MemLimiter) Access(id string) (bool, error) {
	l.mux.Lock()
	defer l.mux.Unlock()

	bucket, found := l.buckets[id]
	if !found {
		if len(l.buckets) >= l.cap {
			return false, errLimiterIsFull
		}
		l.buckets[id] = NewBucket(l.rstCyc, l.rstTokens)
		return true, nil
	}

	switch bucket.Access() {
	case accessNoToken:
		return false, nil
	case accessOk:
		bucket.Minus()
		return true, nil
	default:
		bucket.Reset(l.rstCyc, l.rstTokens)
		return true, nil
	}
}

func (l *MemLimiter) GetSize() int {
	l.mux.Lock()
	defer l.mux.Unlock()
	return len(l.buckets)
}

func (l *MemLimiter) SetCap(cap int) {
	l.mux.Lock()
	defer l.mux.Unlock()
	if cap > l.cap {
		l.cap = cap
	}
}

func (l *MemLimiter) GetCap() int {
	l.mux.Lock()
	defer l.mux.Unlock()
	return l.cap
}

func (l *MemLimiter) SetResetCyc(rstCyc int64) {
	l.mux.Lock()
	defer l.mux.Unlock()
	l.rstCyc = rstCyc
}

func (l *MemLimiter) GetResetCyc() int64 {
	l.mux.Lock()
	defer l.mux.Unlock()
	return l.rstCyc
}

func (l *MemLimiter) SetRstTokens(rstTokens int) {
	l.mux.Lock()
	defer l.mux.Unlock()
	l.rstTokens = rstTokens
}

func (l *MemLimiter) GetRstTokens() int {
	l.mux.Lock()
	defer l.mux.Unlock()
	return l.rstTokens
}

// auto clean func