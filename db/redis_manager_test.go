package db

import (
	"context"
	"liveChat/entities"
	"reflect"
	"strconv"
	"testing"
	"time"
)

const testRedisConfigFilePath = "../default_config_files/default_redis_config.json"

func TestInitRedisConnection(t *testing.T) {
	if err := InitRedisConnection(testRedisConfigFilePath); err != nil {
		t.Fatalf("Redis connection initaition failed: %s", err.Error())
	}
}

func TestRedisLockAndUnLock(t *testing.T) {
	redisConnection.FlushAll(context.Background())
	lock := NewRedisLock(strconv.FormatInt(12345, 10), 10)

	if ok, _ := lock.Unlock(); ok {
		t.Fatalf("Redis lock initially locked")
	}

	if ok, err := lock.Lock(); err != nil {
		t.Fatalf("Redis lock failed: %s", err.Error())
	} else if !ok {
		t.Fatalf("Redis lock has locked")
	}

	if ok, err := lock.Lock(); err != nil {
		t.Fatalf("Redis lock re-entry throw eroor: %s", err.Error())
	} else if ok {
		t.Fatalf("Redis lock re-entry unexpectedly succeeded")
	}

	time.Sleep(redisLockTimeOut + 4*time.Second)

	if ok, err := lock.Unlock(); err != nil {
		t.Fatalf("Redis lock unlock failed: %s", err.Error())
	} else if !ok {
		t.Fatalf("Redis lock has not locked yet")
	}
}

func TestRedisLockMulti(t *testing.T) {
	lock1 := NewRedisLock(strconv.FormatInt(12345, 10), 10)
	lock2 := NewRedisLock(strconv.FormatInt(12345, 10), 11)

	if _, err := lock1.Lock(); err != nil {
		t.Fatalf("Redis Lock1 lock failed: %s", err.Error())
	}

	if ok, _ := lock2.Lock(); ok {
		t.Fatalf("Redis Lock2 lock unexpectedly succeeded")
	}

	if ok, err := lock1.Unlock(); err != nil {
		t.Fatalf("Redis Lock1 unlock failed: %s", err.Error())
	} else if !ok {
		t.Fatalf("Redis Lock1 has not locked yet")
	}

	if ok, err := lock2.Lock(); err != nil {
		t.Fatalf("Redis Lock2 lock failed: %s", err.Error())
	} else if !ok {
		t.Fatalf("Redis Lock2 has not locked yet")
	}

	if ok, err := lock2.Unlock(); err != nil {
		t.Fatalf("Redis Lock2 unlock failed: %s", err.Error())
	} else if !ok {
		t.Fatalf("Redis Lock2 has not locked yet")
	}
}

func TestCacheAndFetchMessageCache(t *testing.T) {
	msg := entities.NewMessage(12, 1, 2, 0, entities.Text, "测试字符串")
	if err := CacheMessageWithTimeOut(msg); err != nil {
		t.Fatalf("Redis cache message failed: %s", err.Error())
	}
	if cacheMsg, err := FetchMessageCache(2, 12); err != nil {
		t.Fatalf("Redis fetch message failed: %s", err.Error())
	} else if !reflect.DeepEqual(cacheMsg, msg) {
		t.Fatalf("Redis cache message mismacthed origin message. origin: %+v. received: %+v", msg, cacheMsg)
	}
}
