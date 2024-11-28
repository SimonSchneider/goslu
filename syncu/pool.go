package syncu

import "sync"

type Pool[T any] struct {
	inner sync.Pool
}

func NewPool[T any](new func() T) *Pool[T] {
	return &Pool[T]{inner: sync.Pool{New: func() any { return new() }}}
}

func (p *Pool[T]) Get() T {
	return p.inner.Get().(T)
}

func (p *Pool[T]) Put(x T) {
	p.inner.Put(x)
}
