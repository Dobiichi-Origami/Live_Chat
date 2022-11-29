package pool

import (
	"liveChat/tcp"
	"sync"
)

var ctxPool sync.Pool

func init() {
	ctxPool = sync.Pool{New: func() interface{} {
		return &tcp.TCPContext{}
	}}
}

func GetTCPContext() interface{} {
	return slicePool.Get()
}

func PutTCPContext(ctx interface{}) {
	slicePool.Put(ctx)
}
