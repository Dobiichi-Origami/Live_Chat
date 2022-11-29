package db

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	"go.uber.org/atomic"
	"liveChat/config"
	"liveChat/entities"
	"liveChat/tools"
	"strconv"
	"time"
)

const defaultRedisConfigPath = "./redis_config.json"

var RedisConfigPath = defaultRedisConfigPath

var (
	redisConnection *redis.Client = nil

	tokenTimeOut           = time.Hour
	redisLockTimeOut       = time.Second * 10
	messageCacheTimeOut    = time.Hour
	friendshipCacheTimeOut = time.Hour * 24

	isRedisInitiated = false
)

var RedisNoResultError = errors.New("Redis 内不存在值")

func InitRedisConnection(configPath string) error {
	if isRedisInitiated {
		return nil
	}

	path := tools.GetPath(RedisConfigPath, configPath)
	cfg := config.NewRedisConfig(path)
	url := cfg.Format()

	options, err := redis.ParseURL(url)
	if err != nil {
		return err
	}

	redisConnection = redis.NewClient(options)
	err = redisConnection.Ping(context.Background()).Err()
	if err != nil {
		return err
	}

	isRedisInitiated = true
	return nil
}

func CacheMessageWithTimeOut(m *entities.Message) error {
	key := getCacheMessageKey(m.Receiver, m.Id)
	val, err := m.MarshalJSON()
	if err != nil {
		return err
	}
	return redisConnection.SetEX(context.Background(), key, val, messageCacheTimeOut).Err()
}

func FetchMessageCache(chatId int64, seq uint64) (*entities.Message, error) {
	key := getCacheMessageKey(chatId, seq)
	ret := redisConnection.GetEx(context.Background(), key, messageCacheTimeOut)
	if ret.Err() != nil {
		return nil, returnNilForRedisNil(ret.Err())
	}

	result, err := ret.Result()
	if err != nil {
		return nil, err
	}
	message := entities.NewEmptyMessage()

	if err = message.UnmarshalJSON([]byte(result)); err != nil {
		return nil, err
	}
	return message, nil
}

var (
	luaScriptAtomicCheckAndSetToken = redis.NewScript(luaScriptAtomicCheckAndSetExpiringTxt)
	luaScriptAtomicCheckToken       = redis.NewScript(luaScriptAtomicCheckAndResetTxt)
)

func RedisSetAndCheckTimeoutToken(token string, userId int64) (string, error) {
	cmd := luaScriptAtomicCheckAndSetToken.Run(context.Background(), redisConnection, []string{token}, userId, tokenTimeOut/time.Second)
	return cmd.String(), cmd.Err()
}

func RedisCheckAndResetToken(token string) (int64, error) {
	result, err := luaScriptAtomicCheckToken.Run(context.Background(), redisConnection, []string{token}, tokenTimeOut/time.Second).Int64()
	if err != nil {
		return -1, err
	}
	return result, nil
}

var (
	luaScriptAtomicSetGroupCache = redis.NewScript(luaScriptAtomicSetGroupCacheTxt)
)

func SetGroupInfoCache(info *entities.GroupInfo) (string, error) {
	updateTime := info.UpdatedAt.UnixMilli()
	data, err := json.Marshal(info)
	if err != nil {
		return "", err
	}
	md5Buf := md5.Sum(data)
	md5Ret := string(md5Buf[:])

	err = luaScriptAtomicSetGroupCache.Run(
		context.Background(),
		redisConnection,
		[]string{md5Ret, strconv.FormatInt(updateTime, 10), strconv.FormatInt(info.Id, 10)},
		data).Err()
	if err != nil {
		return "", err
	}
	return md5Ret, nil
}

func CheckAndCmpGroupInfoCache(groupId int64, md5Val string) (bool, error) {
	ret := redisConnection.HGet(context.Background(), "groupInfoHash", strconv.FormatInt(groupId, 10))
	if ret.Err() != nil {
		return false, returnNilForRedisNil(ret.Err())
	}

	return md5Val == ret.String(), nil
}

func PullGroupInfoCache(groupId int64) (*entities.GroupInfo, error) {
	ret := redisConnection.HGet(context.Background(), "groupInfo", strconv.FormatInt(groupId, 10))
	if ret.Err() != nil {
		return nil, returnNilForRedisNil(ret.Err())
	}

	info := &entities.GroupInfo{}
	err := json.Unmarshal([]byte(ret.String()), info)
	if err != nil {
		return nil, err
	}
	return info, nil
}

var (
	luaScriptAtomicSetFriendshipCache = redis.NewScript(luaScriptAtomicSetFriendshipCacheTxt)
)

func SetFriendshipCache(userKey string, updateTime int64, isFriend bool) error {
	return luaScriptAtomicSetFriendshipCache.Run(context.Background(),
		redisConnection,
		[]string{userKey, strconv.FormatInt(updateTime, 10)},
		isFriend).Err()
}

func PullFriendshipCache(userKey string) (bool, error) {
	ret := redisConnection.HGet(context.Background(), "friendship", userKey)
	if ret.Err() != nil {
		return false, ret.Err()
	}

	return ret.Bool()
}

const (
	LockSuccess = iota
	LockFailed
	LockNotFound
)

var (
	luaScriptAtomicLockChatId   = redis.NewScript(luaScriptAtomicLockChatIdTxt)
	luaScriptAtomicUnlockChatId = redis.NewScript(luaScriptAtomicUnlockChatIdTxt)
	luaScriptAtomicReLockChatId = redis.NewScript(luaScriptAtomicReLockChatIdTxt)
)

const defaultDeleteDelay = 3

type RedisLock struct {
	lockId        string
	lockName      string
	isLocked      *atomic.Bool
	retryError    *atomic.Error
	internalTimer *time.Ticker
	unlockChannel chan struct{}
}

func NewRedisLock(lockName string, machineId int64) *RedisLock {
	return &RedisLock{
		lockId: strconv.FormatInt(machineId, 10) + "_" +
			strconv.FormatInt(tools.GenerateSnowflakeId(false), 10),
		lockName:      lockName + "_lock",
		isLocked:      atomic.NewBool(false),
		retryError:    atomic.NewError(nil),
		internalTimer: nil,
		unlockChannel: make(chan struct{}, 1),
	}
}

func (lock *RedisLock) Lock() (bool, error) {
	if lock.isLocked.Load() {
		return false, nil
	}

	// lua 脚本的执行原子性保证了获取锁的原子性
	if ok, err := checkLockReturnValue(luaScriptAtomicLockChatId.Run(context.Background(), redisConnection, []string{lock.lockName}, lock.lockId, redisLockTimeOut/time.Second+defaultDeleteDelay)); !ok {
		return false, err
	}

	lock.internalTimer = time.NewTicker(redisLockTimeOut)
	go func() {
		for {
			select {
			case <-lock.internalTimer.C:
				if ok, err := checkLockReturnValue(luaScriptAtomicReLockChatId.Run(context.Background(), redisConnection, []string{lock.lockName}, lock.lockId, redisLockTimeOut/time.Second+defaultDeleteDelay)); !ok {
					lock.retryError.Store(err)
					lock.isLocked.Store(false)
					return
				}

			case <-lock.unlockChannel:
				return
			}
		}

	}()

	lock.isLocked.Store(true)
	return true, nil
}

func (lock *RedisLock) Unlock() (bool, error) {
	if !lock.isLocked.Load() {
		return false, lock.retryError.Load()
	}

	defer func() {
		lock.unlockChannel <- struct{}{}
		lock.isLocked.Store(false)
	}()

	if ok, err := checkLockReturnValue(luaScriptAtomicUnlockChatId.Run(context.Background(), redisConnection, []string{lock.lockName}, lock.lockId)); !ok {
		return false, err
	}

	return true, nil
}

func checkLockReturnValue(result *redis.Cmd) (bool, error) {
	val, err := result.Int()
	if err != nil {
		return false, err
	} else if val != LockSuccess {
		// TODO 封装返回枚举值不匹配错误
		return false, nil
	}

	return true, nil
}

func getCacheMessageKey(chatId int64, seq uint64) string {
	return strconv.FormatInt(chatId, 10) + "_" + strconv.FormatUint(seq, 10)
}

func returnNilForRedisNil(err error) error {
	if err == redis.Nil {
		return RedisNoResultError
	}
	return err
}

var (
	luaScriptAtomicLockChatIdTxt = `
local lockName = KEYS[1]
local lockId = ARGV[1]
local timeout = ARGV[2]

local ret = redis.call("EXISTS", lockName)
if ret == 1 then
    return 1
end

redis.call("SETEX", lockName, timeout, lockId)
return 0`

	luaScriptAtomicUnlockChatIdTxt = `
local lockName = KEYS[1]
local lockId = ARGV[1]

local ret = redis.call("EXISTS", lockName)
if ret == 0 then
    return 2
end

ret = redis.call("GET", lockName)
if ret == lockId then
    redis.call("DEL", lockName)
    return 0
end

return 1`

	luaScriptAtomicReLockChatIdTxt = `
local lockName = KEYS[1]
local lockId = ARGV[1]
local timeout = ARGV[2]

local ret = redis.call("EXISTS", lockName)
if ret == 0 then
    return 2
end

ret = redis.call("GET", lockName)
if ret == lockId then
    redis.call("SETEX", lockName, timeout, lockId)
    return 0
end

return 1`

	luaScriptAtomicCheckAndSetExpiringTxt = `
local token = KEYS[1]
local userIdInt = ARGV[1]
local userId = tostring(ARGV[1])
local timeOut = ARGV[2]

local ret = redis.call("EXISTS", userId)
if ret then
token = redis.call("GETEX", userId, "EX", timeOut)
redis.call("SETEX", token, timeOut, userIdInt)
return token
end

redis.call("SETEX", token, timeOut, userIdInt)
redis.call("SETEX", userId, timeOut, token)
return token`

	luaScriptAtomicCheckAndResetTxt = `
local token = KEYS[1]
local timeOut = ARGV[2]

local ret = redis.call("EXISTS", token)
if not ret then
return -1
end

local userIdInt = redis.call("GETEX", token, "EX", timeOut)
redis.call("SETEX", tostring(userIdInt), timeOut, token)

return userIdInt`

	luaScriptAtomicSetGroupCacheTxt = `
local md5 = KEYS[1]
local updateTime = KEYS[2]
local groupId = KEYS[3]
local memberIdJson = ARGV[1]

local updateTimeTableName = "groupInfoUpdateTime"
local groupInfoTableName = "groupInfo"
local hashTableName = "groupInfoHash"

local ret = redis.call("HGET", updateTimeTableName, groupId)
if (not ret) or (ret < updateTime) then
    redis.call("HSET", updateTimeTableName, groupId, updateTime)
    redis.call("HSET", hashTableName, groupId, md5)
    redis.call("HSET", groupInfoTableName, groupId, memberIdJson)
end`

	luaScriptAtomicSetFriendshipCacheTxt = `
local userKey = KEYS[1]
local updateTime = KEYS[2]
local isFriend = ARGV[1]

local updateTimeTableName = "friendshipUpdateTime"
local friendshipTableName = "friendship"

local ret = redis.call("HGET", updateTimeTableName, userKey)
if (not ret) or (ret < updateTime) then
    redis.call("HSET", updateTimeTableName, userKey, updateTime)
    redis.call("HSET", friendshipTableName, userKey, isFriend)
end`
)
