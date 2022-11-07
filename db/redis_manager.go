package db

import (
	"context"
	"fmt"
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

	ChatIdTimeOut       = time.Hour * 30
	ChatIdLockTimeOut   = time.Second * 10
	MessageCacheTimeOut = time.Hour * 1

	isRedisInitiated = false
)

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
	if cmd := redisConnection.SetEX(context.Background(), key, val, MessageCacheTimeOut); cmd.Err() != nil {
		return cmd.Err()
	}
	return nil
}

func FetchMessageCache(chatId int64, seq uint64) (*entities.Message, error) {
	key := getCacheMessageKey(chatId, seq)
	ret := redisConnection.GetEx(context.Background(), key, MessageCacheTimeOut)
	if ret.Err() != nil {
		return nil, ret.Err()
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

func getCacheMessageKey(chatId int64, seq uint64) string {
	return strconv.FormatInt(chatId, 10) + "_" + strconv.FormatUint(seq, 10)
}

const (
	luaScriptAtomicSetAndIncrChatSeqTxt = `
local chatId = KEYS[1]
local seq = ARGV[1]

local ret = redis.call("EXISTS", chatId)
if not ret then
	redis.call("SETEX", chatId, %d, seq+1)
else
	seq = redis.call("GETEX", chatId, "EX", %d)
	redis.call("INCR", chatId)
end

return seq
`

	luaScriptAtomicLockChatIdTxt = `
local lockName = KEYS[1]
local lockId = ARGV[1]

local ret = redis.call("EXISTS", lockName)
if ret == 1 then
	return 1
end

redis.call("SETEX", lockName, %d, lockId)
return 0
`

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

return 1
`

	luaScriptAtomicReLockChatIdTxt = `
local lockName = KEYS[1]
local lockId = ARGV[1]

local ret = redis.call("EXISTS", lockName)
if ret == 0 then
	return 2
end

ret = redis.call("GET", lockName)
if ret == lockId then
	redis.call("SETEX", lockName, %d, lockId)
	return 0
end

return 1
`
)

const (
	LockSuccess = iota
	LockFailed
	LockNotFound
)

var (
	luaScriptAtomicSetAndIncrChatSeq = redis.NewScript(fmt.Sprintf(luaScriptAtomicSetAndIncrChatSeqTxt,
		ChatIdTimeOut/time.Second, ChatIdTimeOut/time.Second))

	luaScriptAtomicLockChatId   = redis.NewScript(fmt.Sprintf(luaScriptAtomicLockChatIdTxt, ChatIdLockTimeOut/time.Second+defaultDeleteDelay))
	luaScriptAtomicUnlockChatId = redis.NewScript(luaScriptAtomicUnlockChatIdTxt)
	luaScriptAtomicReLockChatId = redis.NewScript(fmt.Sprintf(luaScriptAtomicReLockChatIdTxt, ChatIdLockTimeOut/time.Second+defaultDeleteDelay))
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

func NewRedisLock(lockId, machineId int64) *RedisLock {
	return &RedisLock{
		lockId: strconv.FormatInt(machineId, 10) + "_" +
			strconv.FormatInt(tools.GenerateSnowflakeId(false), 10),
		lockName:      strconv.FormatInt(lockId, 10) + "_lock",
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
	if ok, err := checkLockReturnValue(luaScriptAtomicLockChatId.Run(context.Background(), redisConnection, []string{lock.lockName}, lock.lockId)); !ok {
		return false, err
	}

	lock.internalTimer = time.NewTicker(ChatIdLockTimeOut)
	go func() {
		for {
			select {
			case <-lock.internalTimer.C:
				if ok, err := checkLockReturnValue(luaScriptAtomicReLockChatId.Run(context.Background(), redisConnection, []string{lock.lockName}, lock.lockId)); !ok {
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
