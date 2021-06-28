local json = require 'json'
local uuid = require 'uuid'
--local sBox = require 'sBox'

box.cfg {
    work_dir = "/data",
    listen = 3301,
    slab_alloc_arena = 0.256,
    snapshot_period = 3600
}

local function rendSuccess(req)
    return req:render { json = { status = 'success'} }
end

local function rendSuccessPaiload(req, data)
    return req:render { json = { status = 'success', data = data} }
end

local function rendError(req, message, code)
    return req:render { json = { status = 'error', message = message, code = code } }
end

local function tuple2Json(tuple)
    return {
        token = tuple[sBox.col.token],
        user_id = tuple[sBox.col.userId],
        create = tuple[sBox.col.create],
        activity = tuple[sBox.col.activity],
        ip = tuple[sBox.col.ip],
    }
end

local function new_auth_token(req)
    local user_uuid, token, pin_code, phone = req:param('user_uuid'), req:param('token')
    if (not user_uuid or not uuid.is_uuid(user_uuid)) then return rendError(req, 'invalid user uuid', 'invalid_user_uuid') end
    if (not token or token == '') then return rendError(req, 'invalid token', 'invalid_token') end
    if box.tuple.is(sBox.space:insert{token, uuid.fromstr(user_uuid), pin_code, phone, 0 }) then return rendSuccess(req) end
    rendError(req, 'error create auth token', 'err_create_auth_token')
end

--local function get(req)
--    local token, ip, extra = req:param('token'), req:param('ip'), req:post_param() -- ip and extra is optional
--    if (not token) then return rendError(req, 'token not found', 'token_not_found'); end
--
--    local updateData = { { '=', sBox.col.activity, os.time() } }
--    if next(extra) then table.insert(updateData, { '=', sBox.col.extra, post2Extra(extra) }) end
--    if (ip and ip ~= '') then table.insert(updateData, { '=', sBox.col.ip, ip }) end
--
--    local tuple = sBox.space:update(token, updateData)
--    if (not tuple) then return rendError(req, 'token not found', 'token_not_found'); end
--
--    return rendSuccess(req, tuple2Json(tuple))
--end
--
--local function del(req)
--    local token = req:param('token')
--    if (not token) then return rendError(req, 'token not found', 'token_not_found'); end
--    local tuple = sBox.space:delete(token)
--    if (not tuple) then return rendError(req, 'token not found', 'token_not_found'); end
--    return rendSuccess(req, tuple2Json(tuple))
--end
--
--local function user(req)
--    local userId = req:param('id')
--    if (not tonumber(userId)) then return rendError(req, 'invalid user id', 'invalid_user_id') end
--    local tuples = sBox.space.index[sBox.index.userId]:select({ math.floor(userId) }, { iterator = 'REQ' })
--    local res = {}
--    for i = 1, #tuples, 1 do table.insert(res, tuple2Json(tuples[i])) end
--    return rendSuccess(req, res)
--end
--
--local function ip(req)
--    local ip = req:param('ip')
--    if (not ip) then return rendError(req, 'invalid ip', 'invalid_ip') end
--    local tuples = sBox.space.index[sBox.index.ip]:select({ ip }, { iterator = 'REQ' })
--    local res = {}
--    for i = 1, #tuples, 1 do table.insert(res, tuple2Json(tuples[i])) end
--    return rendSuccess(req, res)
--end

return {
    new_auth_token = new_auth_token,
    --get = get,
    --del = del,
    --user = user,
    --ip = ip
}