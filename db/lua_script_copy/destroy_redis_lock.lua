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