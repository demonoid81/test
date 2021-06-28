package users

import (
	"github.com/google/uuid"
)

type Token struct {
	Token    *string   `json:"token"`
	UUIDUser uuid.UUID `json:"user_uuid,omitempty"`
	UserType string    `json:"user_type,omitempty"`
	PinCode  string    `json:"pin_code,omitempty"`
	Phone    string    `json:"phone,omitempty"`
	Expire   uint64    `json:"expire,omitempty"`
}

type TokenTarantool struct {
	Status  string  `json:"status"`
	Tokens  []Token `json:"data,omitempty"`
	Message string
	Code    string
}