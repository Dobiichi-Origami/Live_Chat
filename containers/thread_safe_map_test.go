package containers

import (
	"liveChat/tools"
	"strconv"
	"sync"
	"testing"
)

const testElementScale = 100000

var (
	idSlice       = make([]int64, testElementScale, testElementScale)
	stringIdSlice = make([]string, testElementScale, testElementScale)

	tank    = NewThreadSafeContainer()
	conMap  = New()
	syncMap = sync.Map{}
	cmpMap  = make(map[int64]interface{})
)

func TestMain(m *testing.M) {
	for i := 0; i < testElementScale; i++ {
		idSlice[i] = tools.GenerateSnowflakeId(false)
		stringIdSlice[i] = strconv.FormatInt(idSlice[i], 10)
	}
	m.Run()
}

func BenchmarkThreadSafeContainer_Set(b *testing.B) {
	for i := 0; i < testElementScale; i++ {
		tank.Set(idSlice[i], i)
	}
}

func BenchmarkConcurrentMap_Set(b *testing.B) {
	for i := 0; i < testElementScale; i++ {
		conMap.Set(stringIdSlice[i], i)
	}
}

func BenchmarkSyncMap_Set(b *testing.B) {
	for i := 0; i < testElementScale; i++ {
		syncMap.Store(idSlice[i], i)
	}
}

func BenchmarkMap_Set(b *testing.B) {
	for i := 0; i < testElementScale; i++ {
		cmpMap[idSlice[i]] = i
	}
}

func BenchmarkThreadSafeContainer_Get(b *testing.B) {
	for i := 0; i < testElementScale; i++ {
		if ret := tank.Get(idSlice[i]); ret != i {
			b.Fatalf("Thread safe container get wrong data in index %d. Data: %d", i, ret)
		}
	}
}

func BenchmarkConcurrentMap_Get(b *testing.B) {
	for i := 0; i < testElementScale; i++ {
		if ret, _ := conMap.Get(stringIdSlice[i]); ret != i {
			b.Fatalf("concurrent map get wrong data in index %d. Data: %d", i, ret)
		}
	}
}

func BenchmarkSyncMap_Get(b *testing.B) {
	for i := 0; i < testElementScale; i++ {
		if ret, _ := syncMap.Load(idSlice[i]); ret != i {
			b.Fatalf("golang map get wrong data in index %d. Data: %d", i, ret)
		}
	}
}

func BenchmarkMap_Get(b *testing.B) {
	for i := 0; i < testElementScale; i++ {
		if ret := cmpMap[idSlice[i]]; ret != i {
			b.Fatalf("golang map get wrong data in index %d. Data: %d", i, ret)
		}
	}
}

//func BenchmarkThreadSafeContainer_Get(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		if ret := tank.Get(idSlice[i]).(int); ret != i {
//			b.Errorf("Thread safe container benchmark failed in index %d. get val: %d", i, ret)
//		}
//	}
//}
