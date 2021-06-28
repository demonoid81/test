package models

import (
	"fmt"
	"io"
	"strconv"
)

type WorkerOrder string

const (
	WorkerOrderPrimary   WorkerOrder = "Primary"
	WorkerOrderSecondary WorkerOrder = "Secondary"
)

func (e WorkerOrder) IsValid() bool {
	switch e {
	case WorkerOrderPrimary, WorkerOrderSecondary:
		return true
	}
	return false
}

func (e WorkerOrder) Point() *WorkerOrder {
	return &e
}

func (e WorkerOrder) String() string {
	return string(e)
}

func (e *WorkerOrder) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = WorkerOrder(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid WorkerOrder", str)
	}
	return nil
}

func (e WorkerOrder) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
