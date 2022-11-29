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
end

