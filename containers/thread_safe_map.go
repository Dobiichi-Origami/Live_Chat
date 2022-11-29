package containers

import (
	"github.com/cespare/xxhash/v2"
	"go.uber.org/atomic"
	"runtime"
	"sync"
	"unsafe"
)

type ThreadSafeContainer struct {
	iContainer0 *internalContainer
	iContainer1 *internalContainer

	isOping *atomic.Bool
	opLock  *sync.RWMutex
	ptrLock *sync.RWMutex
}

func NewThreadSafeContainer() *ThreadSafeContainer {
	sc := &ThreadSafeContainer{
		iContainer0: newInternalContainer(minimumInitialBucketNumber),
		iContainer1: nil,
		isOping:     atomic.NewBool(false),
		opLock:      &sync.RWMutex{},
		ptrLock:     &sync.RWMutex{},
	}
	go sc.maintainer()
	return sc
}

func (sc *ThreadSafeContainer) Set(key int64, val interface{}) {
	sc.lock()
	defer sc.unlock()

	if !sc.isOping.Load() {
		sc.iContainer0.setKVInBucket(key, val)
		return
	}

	if !sc.iContainer0.setKVInBucket(key, val) {
		sc.iContainer1.setKVInBucket(key, val)
	}
}

func (sc *ThreadSafeContainer) Get(key int64) (interface{}, bool) {
	sc.lock()
	defer sc.unlock()

	if !sc.isOping.Load() {
		return sc.iContainer0.getValueFromBucket(key)
	}

	ret, ok := sc.iContainer0.getValueFromBucket(key)
	if ok {
		return ret, true
	}

	return sc.iContainer1.getValueFromBucket(key)
}

func (sc *ThreadSafeContainer) Delete(key int64) bool {
	sc.lock()
	defer sc.unlock()

	if !sc.isOping.Load() {
		return sc.iContainer0.deleteKVFromBucket(key)
	}

	if ok := sc.iContainer0.deleteKVFromBucket(key); ok {
		return true
	}
	return sc.iContainer1.deleteKVFromBucket(key)
}

func (sc *ThreadSafeContainer) LoadOrStore(key int64, val interface{}) (interface{}, bool) {
	if val == nil {
		return nil, false
	}

	sc.lock()
	defer sc.unlock()

	if !sc.isOping.Load() {
		return sc.iContainer0.loadOrStore(key, val)
	}

	if ret, ok := sc.iContainer0.loadOrStore(key, val); ret != nil {
		return ret, ok
	}
	return sc.iContainer1.loadOrStore(key, val)
}

func (sc *ThreadSafeContainer) lock() {
	sc.ptrLock.RLock()
	sc.opLock.RLock()
}

func (sc *ThreadSafeContainer) unlock() {
	sc.opLock.RUnlock()
	sc.ptrLock.RUnlock()
}

func (sc *ThreadSafeContainer) maintainer() {
	for {
		op := <-sc.iContainer0.signal
		newBucketNumber := calculateNewBucketNumber(op, sc.iContainer0.size.Load())
		sc.iContainer1 = newInternalContainer(newBucketNumber)
		sc.iContainer1.opFlag.Store(op)

		sc.opLock.Lock()
		sc.isOping.Store(true)
		sc.opLock.Unlock()

		rehashDataToNewBucket(sc.iContainer0, sc.iContainer1)

		sc.ptrLock.Lock()
		sc.iContainer0 = sc.iContainer1
		sc.iContainer1 = nil
		sc.iContainer0.opFlag.Store(null)
		sc.isOping.Store(false)
		sc.ptrLock.Unlock()
	}
}

func calculateNewBucketNumber(op int32, oldCap uint64) uint64 {
	newCap := uint64(0)
	if op == expansion {
		//if oldCap >= pow2ExpansionLimitSize {
		//	newCap = oldCap + limitExpansionCapacity
		//} else {
		//	newCap = oldCap * 2
		//}
		newCap = oldCap << 1
	} else if op == shrink {
		newCap = oldCap / 2
		//if newCap < minimumInitialBucketNumber {
		//	newCap = minimumInitialBucketNumber
		//}
	}
	return newCap
}

func rehashDataToNewBucket(oldBucket, newBucket *internalContainer) {
	wg := sync.WaitGroup{}
	for i := 0; i < runtime.NumCPU(); i++ {
		baseStep := len(oldBucket.buckets) / runtime.NumCPU()
		i := i
		wg.Add(1)
		go func() {
			for j := baseStep * i; j < baseStep*(i+1) || (i == 9 && j < len(oldBucket.buckets)); j++ {
				oldBucket.rwLockLists[j].Lock()
				for k := 0; k < len(oldBucket.buckets[j]); k++ {
					newBucket.setKVInBucket(oldBucket.buckets[j][k].key, oldBucket.buckets[j][k].val)
				}
				oldBucket.flipBucketMovedFlag(uint64(j))
				oldBucket.buckets[j] = nil
				oldBucket.rwLockLists[j].Unlock()
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

const (
	minimumInitialBucketNumber = 1 << 4
	pow2ExpansionLimitSize     = 1 << 20
	limitExpansionCapacity     = 1 << 18

	expansionLoadFactor         = 6.5
	shrinkLoadFactor            = 0.1
	bucketElementOverflowFactor = 1 << 10
)

const (
	null int32 = iota
	expansion
	shrink
)

type internalContainer struct {
	opFlag *atomic.Int32
	signal chan int32

	bucketNumber uint64
	size         *atomic.Uint64

	buckets     [][]element
	moveFlag    []bool
	rwLockLists []sync.RWMutex
}

func newInternalContainer(size uint64) *internalContainer {
	return &internalContainer{
		opFlag:       atomic.NewInt32(null),
		signal:       make(chan int32, 1),
		bucketNumber: size,
		size:         atomic.NewUint64(0),
		buckets:      make([][]element, size, size),
		moveFlag:     make([]bool, size, size),
		rwLockLists:  make([]sync.RWMutex, size, size),
	}
}

func (c *internalContainer) getValueFromBucket(key int64) (interface{}, bool) {
	index := c.getIndex(key)
	c.rwLockLists[index].RLock()
	defer c.rwLockLists[index].RUnlock()

	if c.bucketIsMoved(index) {
		return nil, false
	}

	for _, e := range c.buckets[index] {
		if e.key == key {
			return e.val, true
		}
	}
	return nil, true
}

func (c *internalContainer) setKVInBucket(key int64, val interface{}) bool {
	index := c.getIndex(key)
	c.rwLockLists[index].Lock()
	defer func() {
		c.size.Inc()
		c.calculateLoadFactor(index)
		c.rwLockLists[index].Unlock()
	}()

	if c.bucketIsMoved(index) {
		return false
	}

	for _, e := range c.buckets[index] {
		if e.key == key {
			e.val = val
			return true
		}
	}
	c.buckets[index] = append(c.buckets[index], element{key, val})
	return true
}

func (c *internalContainer) loadOrStore(key int64, val interface{}) (interface{}, bool) {
	index := c.getIndex(key)
	c.rwLockLists[index].Lock()
	defer c.rwLockLists[index].Unlock()

	if c.bucketIsMoved(index) {
		return nil, false
	}

	for _, e := range c.buckets[index] {
		if e.key == key {
			return e.val, true
		}
	}

	c.buckets[index] = append(c.buckets[index], element{key, val})
	c.size.Inc()
	c.calculateLoadFactor(index)
	return val, false
}

func (c *internalContainer) deleteKVFromBucket(key int64) bool {
	index := c.getIndex(key)
	c.rwLockLists[index].Lock()
	defer func() {
		c.size.Dec()
		c.calculateLoadFactor(index)
		c.rwLockLists[index].Unlock()
	}()

	if c.bucketIsMoved(index) {
		return false
	}

	for i, e := range c.buckets[index] {
		if e.key == key {
			c.buckets[index] = append(c.buckets[index][:i], c.buckets[index][i+1:]...)
			return true
		}
	}

	return false
}

func (c *internalContainer) bucketIsMoved(index uint64) bool {
	return c.moveFlag[index]
}

func (c *internalContainer) flipBucketMovedFlag(index uint64) {
	c.moveFlag[index] = true
}

func (c *internalContainer) getBucketNumber() uint64 {
	return c.bucketNumber
}

func (c *internalContainer) getIndex(key int64) uint64 {
	return xxhash.Sum64((*(*[8]byte)(unsafe.Pointer(&key)))[:]) % c.bucketNumber
}

func (c *internalContainer) calculateLoadFactor(index uint64) {
	if c.opFlag.Load() != null {
		return
	}

	factor := float64(c.size.Load() / c.bucketNumber)
	if (factor >= expansionLoadFactor || len(c.buckets[index]) >= bucketElementOverflowFactor) && c.opFlag.CAS(null, expansion) {
		c.signal <- expansion
	} else if factor <= shrinkLoadFactor && c.bucketNumber != minimumInitialBucketNumber && c.opFlag.CAS(null, shrink) {
		c.signal <- shrink
	}
}

type element struct {
	key int64
	val interface{}
}
