package models

import (
	"time"

	"github.com/google/uuid"
)

type MsgStat struct {
	UUIDJob    *uuid.UUID `db:"uuid_job"`
	Job        *Job       `json:"job" relay:"uuid_job" link:"UUIDJob"`
	UUIDPerson *uuid.UUID `db:"uuid_person"`
	Person     *Person    `json:"person" relay:"uuid_person" link:"UUIDPerson"`
	Reading    bool       `json:"reading" db:"reading"`
	Updated    *time.Time `json:"updated" db:"updated"`
}
