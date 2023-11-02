-- 先找找有没有这个限流对象的设置
local val = redis.call('get', KEYS[1])
local expiration = ARGV[1]
local limit = tonumber(ARGV[2])
if val == false then
    -- 不存在 key(过期或为设置)
    if limit < 1 then
        -- 执行限流
        return "true"
    else
        -- set your_service 1 px 100ms
        -- 初始值为 1
        redis.call('set', KEYS[1], 1, 'PX', expiration)
        return "false"
    end
elseif tonumber(val) < limit then
    -- 存在对象且小于阈值
    redis.call('incr', KEYS[1])
    return "false"
else
    -- 大于阈值
    return "true"
end 