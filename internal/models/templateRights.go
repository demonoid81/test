package models

import (
	"fmt"
	"github.com/google/uuid"
	"io"
	"strconv"
	"time"
)

type TemplateRights struct {
	UUID               *uuid.UUID             `json:"uuid" db:"uuid"`
	Created            *time.Time             `json:"created" db:"created"`
	Updated            *time.Time             `json:"updated" db:"updated"`
	IsDeleted          *bool                  `json:"isDeleted" db:"is_deleted"`
	Name               *string                `json:"name"`
	JSONRightsToObject map[string]interface{} `db:"uuid_rights_to_object"`
	RightsToObject     []*RightToObject      `json:"rightsToObject" link:"JSONRightsToObject"`
}

type RightToObject struct {
	Object     *string    `json:"object"`
	Select     *bool      `json:"select"`
	Insert     *bool      `json:"insert"`
	Update     *bool      `json:"update"`
	Delete     *bool      `json:"delete"`
}

type Object string

const (
	ObjectOrganization Object = "Organization"
)

func (e Object) IsValid() bool {
	switch e {
	case ObjectOrganization:
		return true
	}
	return false
}

func (e Object) String() string {
	return string(e)
}

func (e *Object) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = Object(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid Object", str)
	}
	return nil
}

func (e Object) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
