package models

import (
	"fmt"
	"io"
	"strconv"
)

type JobTypeIcon string

const (
	Cleaner JobTypeIcon = "cleaner"
	Cashier JobTypeIcon = "cashier"
	Loader JobTypeIcon = "loader"
	Merchandiser JobTypeIcon = "merchandiser"
)

func (e JobTypeIcon) IsValid() bool {
	switch e {
	case Cleaner,Cashier, Loader, Merchandiser:
		return true
	}
	return false
}

func (e JobTypeIcon) String() string {
	return string(e)
}

func (e *JobTypeIcon) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}
	*e = JobTypeIcon(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid gender", str)
	}
	return nil
}

func (e JobTypeIcon) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
