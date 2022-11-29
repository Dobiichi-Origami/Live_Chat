package pool

import "sync"

var slicePool sync.Pool

func init() {
	slicePool = sync.Pool{New: func() interface{} {
		return make([]byte, 12, 12)
	}}
}

func Get12BytesSlice() []byte {
	return slicePool.Get().([]byte)
}

func Put12BytesSlice(slice []byte) {
	slicePool.Put(slice)
}
