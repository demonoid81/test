package models

import (
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"io"
	"strings"
	"time"
)

// Creates a marshaller which converts a time.Time (date) to a string
func MarshalDate(date time.Time) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, e := io.WriteString(w, fmt.Sprintf("%s%s%s", "\"", date.Format("2006-01-02"), "\""))
		if e != nil {
			panic(e)
		}
	})
}

// Unmarshalls a string to a time.Time (date)
func UnmarshalDate(v interface{}) (time.Time, error) {
	fmt.Println(v)
	str, ok := v.(string)
	if !ok {
		return time.Time{}, fmt.Errorf("date must be strings")
	}
	withoutQuotes := strings.ReplaceAll(str, "\"", "")
	fmt.Println(withoutQuotes)
	if withoutQuotes != "" {
		i, err := time.Parse("2006-01-02", withoutQuotes)
		if err != nil {
			i, err = time.Parse("20060102", withoutQuotes)
		}
		fmt.Println(i.Format("2006-01-02"))
		return i, err
	}
	return time.Time{}, nil
}
