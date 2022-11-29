local lockName = KEYS[1]
local lockId = ARGV[1]
local timeout = ARGV[2]

local ret = redis.call("EXISTS", lockName)
if ret == 1 then
    return 1
end

redis.call("SETEX", lockName, timeout, lockId)
return 0