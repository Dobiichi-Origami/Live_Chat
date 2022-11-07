package pool

import (
	"github.com/panjf2000/ants/v2"
)

type WorkerPool struct {
	*ants.PoolWithFunc
}

func NewWorkerPool(fn func(interface{})) (*WorkerPool, error) {
	tmp, err := ants.NewPoolWithFunc(-1, fn)
	if err != nil {
		return nil, err
	}
	return &WorkerPool{tmp}, nil
}

func (p *WorkerPool) PushTask(arg interface{}) {
	p.Invoke(arg)
}
