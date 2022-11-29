local token = KEYS[1]
local timeOut = ARGV[2]

local ret = redis.call("EXISTS", token)
if not ret then
    return -1
end

local userIdInt = redis.call("GETEX", token, "EX", timeOut)
redis.call("SETEX", tostring(userIdInt), timeOut, token)

return userIdInt