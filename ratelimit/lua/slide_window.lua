-- 1, 2, 3, 4, 5, 6, 7 这是你的元素
-- ZREMRANGEBYSCORE key1 0 6
-- 7 执行完之后

-- 限流对象
local key = KEYS[1]
-- 窗口大小
local window = tonumber(ARGV[1])
-- 最大限制
local threshold = tonumber(ARGV[2])
-- 当前时间戳
local now = tonumber(ARGV[3])

-- 窗口开始时间
local min = now - window

-- 把窗口之前的减去
redis.call('ZREMRANGEBYSCORE', key, '-inf', min)
local cnt = redis.call('ZCOUNT', key, '-inf', '+inf')
-- redis.call('ZCOUNT' key, min, '+inf')
if cnt >= threshold then
    -- 触发限流
    return "true"
else
    -- 把 score 和 member 都设置成 now
    redis.call('ZADD', key, now, now)
    redis.call('PEXPIRE', key, window)
    return "false"
end 