package models

import (
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"io"
	"time"
)

func MarshalDuration(duration time.Duration) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, e := io.WriteString(w, fmt.Sprintf("%s", "\"", duration.String(), "\""))
		if e != nil {
			panic(e)
		}
	})
}

// Unmarshalls a string to a time.Time (time)
func UnmarshalDuration(v interface{}) (time.Duration, error) {
	str, ok := v.(string)
	if !ok {
		return time.Duration(0), fmt.Errorf("time must be strings")
	}
	fmt.Println(str)
	i, err := time.ParseDuration(str)
	return i, err
}
