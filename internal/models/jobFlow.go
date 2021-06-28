package models

import (
	"fmt"
	"github.com/google/uuid"
	"io"
	"strconv"
	"time"
)

type ScriptType string

const (
	ScriptTypePrepare ScriptType = "prepare"
)

var AllScriptType = []ScriptType{
	ScriptTypePrepare,
}

func (e ScriptType) IsValid() bool {
	switch e {
	case ScriptTypePrepare:
		return true
	}
	return false
}

func (e ScriptType) String() string {
	return string(e)
}

func (e *ScriptType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = ScriptType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid ScriptType", str)
	}
	return nil
}

func (e ScriptType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type JobFlow struct {
	UUID         *uuid.UUID    `json:"uuid"`
	Created      *time.Time    `json:"created"`
	Updated      *time.Time    `json:"updated"`
	IsDeleted    *bool         `json:"isDeleted"`
	Organization *Organization `json:"organization"`
	JobType      *JobType      `json:"jobType"`
	ScriptType   *ScriptType   `json:"scriptType"`
	Diff         *int          `json:"diff"`
	Script       *string       `json:"script"`
}
