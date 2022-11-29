package controllers

import (
	"liveChat/containers"
	"liveChat/db"
	"strconv"
	"time"
)

const friendshipUpdateIntervalInMilli = 2000

var (
	friendshipCache          containers.ConcurrentMap
	friendshipCacheTimestamp containers.ConcurrentMap
)

func init() {
	friendshipCache = containers.New()
	friendshipCacheTimestamp = containers.New()
}

func CheckAreUsersFriend(userId1, userId2 int64, isEnforceDb bool) (bool, error) {
	key := getCacheKey(userId1, userId2)
	if err := checkFriendshipCache(key, userId1, userId2, isEnforceDb); err != nil {
		return false, err
	}

	flag, _ := friendshipCache.Get(key)
	return flag.(bool), nil
}

func getCacheKey(userId1, userId2 int64) string {
	if userId1 > userId2 {
		tmp := userId1
		userId1 = userId2
		userId2 = tmp
	}

	return strconv.FormatInt(userId1, 10) + "_" + strconv.FormatInt(userId2, 10)
}

func checkFriendshipCache(key string, userId1, userId2 int64, isEnforceDb bool) error {
	updateFlag := false
	if ret, ok := friendshipCacheTimestamp.Get(key); isEnforceDb || !ok || ret == nil || time.Now().UnixMilli()-ret.(int64) >= friendshipUpdateIntervalInMilli {
		updateFlag = true
	}

	if !updateFlag {
		return nil
	}

	if ret, ok := friendshipCacheTimestamp.Get(key); isEnforceDb || !ok || ret == nil {
		flag, t, err := db.TellIsFriendBetween(nil, userId1, userId2)
		if err != nil {
			return err
		}

		err = db.SetFriendshipCache(key, t, flag)
		if err != nil {
			return err
		}
	}

	flag, err := db.PullFriendshipCache(key)
	if err != nil {
		return err
	}

	friendshipCache.Set(key, flag)
	friendshipCacheTimestamp.Set(key, time.Now().UnixMilli())
	return nil
}
