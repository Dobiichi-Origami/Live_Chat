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
return token