package models

import (
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"io"
	"strings"
	"time"
)

func MarshalTime(time time.Time) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, e := io.WriteString(w, fmt.Sprintf("%s%s%s", "\"", time.Format("15:04:05"), "\""))
		if e != nil {
			panic(e)
		}
	})
}

// Unmarshalls a string to a time.Time (time)
func UnmarshalTime(v interface{}) (time.Time, error) {
	str, ok := v.(string)
	if !ok {
		return time.Time{}, fmt.Errorf("time must be strings")
	}
	fmt.Println(str)
	withoutQuotes := strings.ReplaceAll(str, "\"", "")
	i, err := time.Parse("15:04:05", withoutQuotes)
	if err != nil {
		i, err = time.Parse("15:04:05", withoutQuotes)
	}
	return i.UTC(), err
}