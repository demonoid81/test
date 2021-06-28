package models

import (
	"time"

	"github.com/google/uuid"
)

type Balance struct {
	UUID             *uuid.UUID    `json:"uuid" db:"uuid" auto:"false"`
	Created          *time.Time    `json:"created" db:"created"`
	Updated          *time.Time    `json:"updated" db:"updated"`
	IsDeleted        *bool         `json:"isDeleted" db:"is_deleted"`
	UUIDOrganization *uuid.UUID    `db:"uuid_organization"`
	Organization     *Organization `json:"organization" relay:"uuid_organization" link:"UUIDOrganization"`
	Amount           *float64      `json:"amount" db:"amount"`
	UUIDMovement     *uuid.UUID    `db:"uuid_movement"`
	Movement         *Movement     `json:"movement" relay:"uuid_movement" link:"UUIDMovement"`
}
