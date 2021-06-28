fiber = require('fiber')
uuid = require('uuid')
space = box.schema.create_space('sessions_tokens', { if_not_exists = true } )
space:format({
    {name = 'uuid', type='uuid'},
    {name = 'token', type = 'string'},
    {name = 'user', type = 'uuid', is_nullable = true},
    {name = 'pin_code', type = 'string', is_nullable = true},
    {name = 'phone', type = 'string', is_nullable = true},
    {name = 'expire', type = 'unsigned', is_nullable = true}
}, { if_not_exists = true })

space:create_index('primary', {
    unique = true,
    type = 'hash',
    parts = {1, 'uuid'},
    if_not_exists = true }
)

space:create_index('token', {
    unique = true,
    type = 'hash',
    parts = {2, 'string'},
    if_not_exists = true }
)

space:create_index('user', {
    unique = true,
    type = 'hash',
    parts = {3, 'uuid'},
    if_not_exists = true }
)

space:create_index('pin_code', {
    unique = true,
    type = 'hash',
    parts = {4, 'string'},
    if_not_exists = true }
)

space:create_index('expire', {
    unique = false,
    type = 'tree',
    parts = {6, 'unsigned'},
    if_not_exists = true }
)

expire_loop = fiber.create(
        function ()
            while true do
                local time = math.floor(fiber.time())
                for _, tuple in box.space.sessions_tokens.index['expire']:pairs(time,  { iterator = 'LT' }) do
                    local uuid, token, user, pin_code, phone, expire = tuple:unpack()
                    if expire == 0 then
                        -- временый токен живет всего 3 минуты
                        -- постоянны токен 10 минут
                        if pinCode == "" then
                            box.space.sessions_tokens:replace{uuid, token, user, pin_code, phone, time + 600}
                        else
                            box.space.sessions_tokens:replace{uuid, token, user, pin_code, phone, time + 180000}
                        end
                    else
                        if time > expire then
                            box.space.sessions_tokens:delete{uuid}
                        end
                    end
                end
                fiber.sleep(1)
            end
        end
)

function add_user_session(user, token)
    local sessionsUUID = uuid.new()
    box.space.sessions_tokens:insert{ sessionsUUID, token, uuid.fromstr(user), "", "",0}
end

function add_auth_session(token, user, pin_code, phone)
    local sessionsUUID = uuid.new()
    box.space.sessions_tokens:insert{ sessionsUUID, token, uuid.fromstr(user), pin_code, phone, 0}
end
