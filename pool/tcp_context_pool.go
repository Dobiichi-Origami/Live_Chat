package pool

import (
	"sync"
)

var ctxPool sync.Pool

func init() {
	ctxPool = sync.Pool{New: func() interface{} {
		return &TCPContext{}
	}}
}

func GetTCPContext() interface{} {
	return slicePool.Get()
}

func PutTCPContext(ctx interface{}) {
	slicePool.Put(ctx)
}
