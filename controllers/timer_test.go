package controllers

import (
	"testing"
	"time"
)

func BenchmarkTimer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		last := int64(0)
		for j := 0; j < 100000; j++ {
			now := time.Now().UnixMilli()
			if last == 0 || now-last > 1000 {
				last = now
			}
		}
	}
}

func BenchmarkIteration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		lists := make([]int64, 1000, 1000)
		for i := 0; i < len(lists); i++ {
			if lists[i] == 1 {

			}
		}
	}
}
