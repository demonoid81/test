package models

import (
	"fmt"
	"io"
	"strconv"
)

type ClientType string

const (
	ClientTypeMobile   ClientType = "mobile"
	ClientTypeWeb      ClientType = "web"
	ClientTypeDirector ClientType = "director"
)

var AllClientType = []ClientType{
	ClientTypeMobile,
	ClientTypeWeb,
	ClientTypeDirector,
}

func (e ClientType) IsValid() bool {
	switch e {
	case ClientTypeMobile, ClientTypeWeb, ClientTypeDirector:
		return true
	}
	return false
}

func (e ClientType) String() string {
	return string(e)
}

func (e *ClientType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = ClientType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid ClientType", str)
	}
	return nil
}

func (e ClientType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
