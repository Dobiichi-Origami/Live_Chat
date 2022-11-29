package controllers

import (
	"crypto/md5"
	"encoding/json"
	"liveChat/containers"
	"liveChat/db"
	"liveChat/entities"
	"time"
)

const groupInfoUpdateIntervalInMilli = 3000

var (
	groupInfoCache    *containers.ThreadSafeContainer
	cacheTimestampMap *containers.ThreadSafeContainer
	cacheMd5StringMap *containers.ThreadSafeContainer
)

func init() {
	groupInfoCache = containers.NewThreadSafeContainer()
	cacheTimestampMap = containers.NewThreadSafeContainer()
	cacheMd5StringMap = containers.NewThreadSafeContainer()
}

func CheckIsUserInGroup(userId, groupId int64, isEnforceDb bool) (bool, error) {
	if err := checkGroupInfoCache(groupId, isEnforceDb); err != nil {
		return false, err
	}

	var (
		tmp, _ = groupInfoCache.Get(groupId)
		info   = tmp.(entities.GroupInfo)
	)

	if info.IsDeleted {
		return false, nil
	}

	for _, entry := range info.Members {
		if !entry.IsDeleted && entry.MemberId == userId {
			return true, nil
		}
	}

	return false, nil
}

func GetUserListInGroup(groupId int64, isEnforceDb bool) ([]entities.GroupMember, error) {
	if err := checkGroupInfoCache(groupId, isEnforceDb); err != nil {
		return nil, err
	}

	var (
		tmp, _ = groupInfoCache.Get(groupId)
		info   = tmp.(entities.GroupInfo)
	)

	if info.IsDeleted {
		return nil, nil
	}

	ret := make([]entities.GroupMember, len(info.Members), len(info.Members))
	copy(ret, info.Members)
	return ret, nil
}

func checkGroupInfoCache(groupId int64, isEnforceDb bool) error {
	updateFlag := false
	if ret, ok := cacheTimestampMap.Get(groupId); isEnforceDb || !ok || ret == nil || time.Now().UnixMilli()-ret.(int64) >= groupInfoUpdateIntervalInMilli {
		updateFlag = true
	}

	if !updateFlag {
		return nil
	}

	if ret, ok := cacheTimestampMap.Get(groupId); isEnforceDb || !ok || ret == nil {
		info, err := db.SearchGroupInfo(nil, groupId, true)
		if err != nil {
			return err
		}

		_, err = db.SetGroupInfoCache(info)
		if err != nil {
			return err
		}
	}

	var (
		flag = false
		err  error
	)

	if md5StrLocal, ok := cacheMd5StringMap.Get(groupId); ok {
		flag, err = db.CheckAndCmpGroupInfoCache(groupId, md5StrLocal.(string))
		if err != nil {
			return err
		}
	}

	if !flag {
		info, err := db.PullGroupInfoCache(groupId)
		if err != nil {
			return err
		}

		data, err := json.Marshal(info)
		if err != nil {
			return err
		}

		md5Buf := md5.Sum(data)
		md5Str := string(md5Buf[:])
		cacheMd5StringMap.Set(groupId, md5Str)
		groupInfoCache.Set(groupId, info)
	}

	cacheTimestampMap.Set(groupId, time.Now().UnixMilli())
	return nil
}
