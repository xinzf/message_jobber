package utils

import (
	"sync/atomic"
)

type Pool struct {
	c      chan interface{}
	create func() interface{}

	hit  int64
	miss int64
}

func NewPool(initSize, maxSize int, fun func() interface{}) (p *Pool) {
	if maxSize < 1 {
		maxSize = 1
	}

	if initSize > maxSize {
		initSize = maxSize
	}

	p = &Pool{
		c:      make(chan interface{}, maxSize),
		create: fun,
	}

	if initSize > 0 {
		for i := 0; i < initSize; i++ {
			i := fun()
			// i.Init()
			p.Put(i)
		}
	}

	return
}

func (this *Pool) Get() (o interface{}) {
	select {
	case o = <-this.c:
		atomic.AddInt64(&this.hit, 1)
	default:
		o = this.create()
		atomic.AddInt64(&this.miss, 1)
	}

	// o.Init()
	return
}

func (this *Pool) Put(o interface{}) {
	// o.Reset()

	select {
	case this.c <- o:
	default:
	}
}

func (this *Pool) Hit() int64 {
	return this.hit
}

func (this *Pool) Miss() int64 {
	return this.miss
}
