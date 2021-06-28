local fiber = require 'fiber'

local sBox = {
    col = {
        token = 1,
        user = 2,
        pin_code = 3,
        phone = 4,
        expire = 5
    },
    index = {
        token = 'token',
        user = 'user',
        pin_code = 'pin_code',
        expire = 'expire'
    },
    space = box.schema.create_space('sessions', { if_not_exists = true })
}

box.once('token_index', function()
    sBox.space:create_index(sBox.index.token, { type = 'HASH', parts = { sBox.col.token, 'string' } })
    sBox.space:create_index(sBox.index.user, { unique = user, parts = { sBox.col.user, 'uuid' } })
    sBox.space:create_index(sBox.index.pin_code, { unique = false, parts = { sBox.col.pin_code, 'string' } })
    sBox.space:create_index(sBox.index.expire, { unique = false, parts = { sBox.col.expire, 'unsigned' } })
end)
--box.schema.user.grant('guest', 'read,write,execute', 'universe', nil, { if_not_exists = true })

fiber.create(function ()
    while true do
        local time = math.floor(fiber.time())
        for _, tuple in sBox.space.index[sBox.index.expire]:pairs(time,  { iterator = 'LT' }) do
            local token, user, pin_code, phone, expire = tuple:unpack()
            if expire == 0 then
                -- временый токен живет всего 3 минуты
                -- постоянны токен 10 минут
                if pin_code == "" then
                    sBox.space:replace{token, user, pin_code, phone, time + 600}
                else
                    sBox.space:replace{token, user, pin_code, phone, time + 180000}
                end
            else
                if time > expire then
                    sBox.space:delete{sBox.col.token}
                end
            end
        end
        fiber.sleep(1)
    end
end)

return sBox