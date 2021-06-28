#! /usr/bin/env tarantool
local fiber = require 'fiber'
local json = require 'json'
local uuid = require 'uuid'

box.cfg {
    listen = 3301,
    slab_alloc_arena = 0.256,
    snapshot_period = 3600
}

box.schema.user.passwd('pass')

local sBox = {
    col = {
        token = 1,
        user_uuid = 2,
        user_type = 3,
        pin_code = 4,
        phone = 5,
        expire = 6
    },
    index = {
        token = 'token',
        user_uuid = 'user_uuid',
        pin_code = 'pin_code',
        expire = 'expire'
    },
    space = box.schema.create_space('tokens', { if_not_exists = true })
}

box.once('token_index', function()
    sBox.space:create_index(sBox.index.token, { type = 'HASH', parts = { sBox.col.token, 'string' } })
    sBox.space:create_index(sBox.index.user_uuid, { unique = false, parts = { sBox.col.user_uuid, 'uuid' } })
    sBox.space:create_index(sBox.index.pin_code, { unique = false, parts = { sBox.col.pin_code, 'string' } })
    sBox.space:create_index(sBox.index.expire, { unique = false, parts = { sBox.col.expire, 'unsigned' } })
end)
--box.schema.user.grant('guest', 'read,write,execute', 'universe', nil, { if_not_exists = true })



local function rendSuccess(req)
    return req:render { json = { status = 'success'} }
end

local function rendSuccessPayload(req, data)
    return req:render { json = { status = 'success', data = data} }
end

local function rendError(req, message, code)
    return req:render { json = { status = 'error', message = message, code = code } }
end

local function tuple2Json(tuple)
    return {
        token = tuple[sBox.col.token],
        user_uuid = tuple[sBox.col.user_uuid],
        user_type = tuple[sBox.col.user_type],
        pin_code = tuple[sBox.col.pin_code],
        phone = tuple[sBox.col.phone],
        expire = tuple[sBox.col.expire],
    }
end

local function new_auth_token(req)
    print(json.encode(req:post_param()))
    local user_uuid , token, pin_code, phone = req:post_param('user_uuid'), req:post_param('token'), req:post_param('pin_code'), req:post_param('phone')
    if (not user_uuid) then return rendError(req, 'invalid user uuid', 'invalid_user_uuid') end
    if (not token or token == '') then return rendError(req, 'invalid token', 'invalid_token') end
    if (not user_type) then return rendError(req, 'invalid user type', 'invalid_user_type') end
    if (not pin_code or pin_code == '') then return rendError(req, 'invalid pin code', 'invalid_pin_code') end
    if (not phone or phone == '') then return rendError(req, 'invalid phone number', 'invalid_phone') end
    if box.tuple.is(sBox.space:insert{token, uuid.fromstr(user_uuid), user_type,pin_code, phone, 0 }) then return rendSuccess(req) end
    rendError(req, 'error create auth token', 'err_create_auth_token')
end

local function new_session_token(req)
    print(json.encode(req:post_param()))
    local user_uuid, token, phone = req:post_param('user_uuid'), req:post_param('token'), req:post_param('phone')
    if (not user_uuid) then return rendError(req, 'invalid user uuid', 'invalid_user_uuid') end
    if (not user_type) then return rendError(req, 'invalid user type', 'invalid_user_type') end
    if (not token or token == '') then return rendError(req, 'invalid token', 'invalid_token') end
    if (not phone or phone == '') then return rendError(req, 'invalid phone number', 'invalid_phone') end
    if box.tuple.is(sBox.space:insert{token, uuid.fromstr(user_uuid), user_type, "", phone, 0 }) then return rendSuccess(req) end
    rendError(req, 'error create session token', 'err_create_session_token')
end

local function get_by_token(req)
    local token = req:query_param('token')
    if (not token or token == '') then return rendError(req, 'invalid token', 'invalid_token') end
    local tuples = sBox.space.index[sBox.index.token]:select({ token }, { iterator = 'EQ' })
    local res = {}
    for i = 1, #tuples, 1 do table.insert(res, tuple2Json(tuples[i])) end
    return rendSuccessPayload(req, res)
end

local function get_by_pin_code(req)
    local pin_code = req:query_param('pin_code')
    print(pin_code)
    if (not pin_code or pin_code == '') then return rendError(req, 'invalid pin code', 'invalid_pin_code') end
    local tuples = sBox.space.index[sBox.index.pin_code]:select({ pin_code }, { iterator = 'EQ' })
    local res = {}
    for i = 1, #tuples, 1 do table.insert(res, tuple2Json(tuples[i])) end
    for i = 1, #tuples, 1 do sBox.space:delete(tuples[i][sBox.col.token]) end
    return rendSuccessPayload(req, res)
end

local function del(req)
    local token = req:param('token')
    if (not token) then return rendError(req, 'token not found', 'token_not_found'); end
    local tuple = sBox.space:delete(token)
    if (not tuple) then return rendError(req, 'token not found', 'token_not_found'); end
    return rendSuccess(req)
end


local server = require('http.server').new('0.0.0.0', 80)
server:route({ method = 'POST', path = 'token/auth' }, new_auth_token)
server:route({ method = 'POST', path = 'token/session' }, new_session_token)
server:route({ method = 'GET', path = 'token/getByToken' }, get_by_token)
server:route({ method = 'GET', path = 'token/getByPinCode' }, get_by_pin_code)
server:route({ method = 'DELETE', path = 'token/delete' }, del)
--server:route({ path = '/user' }, ctrl.user)
--server:route({ path = '/ip' }, ctrl.ip)
server:start()