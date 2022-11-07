package controllers

import (
	"go.uber.org/atomic"
	"liveChat/constants"
	"liveChat/containers"
	"time"
)

var (
	timerContainer *containers.ThreadSafeContainer
	localTime      *atomic.Int32
)

func init() {
	timerContainer = containers.NewThreadSafeContainer()
	localTime = atomic.NewInt32(0)
	go func() {
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-ticker.C:
				localTime.Inc()
			}
		}
	}()
}

func NewTimer(fd int) {
	timerContainer.Set(int64(fd), localTime.Load())
}

func CheckTimer(fd int) bool {
	now := localTime.Load()
	if last := timerContainer.Get(int64(fd)); last == nil {
		return false
	} else if now-last.(int32) > constants.HeartBeatMaxInterval {
		return false
	}
	return true
}

func ResetTimer(fd int) {
	if last := timerContainer.Get(int64(fd)); last == nil {
		return
	}
	timerContainer.Set(int64(fd), localTime.Load())
}

func DeleteTimer(fd int) {
	timerContainer.Delete(int64(fd))
}
