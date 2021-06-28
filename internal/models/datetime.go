package models

import (
	"encoding/json"
	"io"
	"log"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/pkg/errors"
)

func MarshalDateTime(t time.Time) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, err := io.WriteString(w, strconv.FormatInt(t.UTC().Unix(), 10))
		if err != nil {
			log.Print(errors.Wrap(err, "while writing marshalled timestamp"))
			return
		}
	})
}

func UnmarshalDateTime(v interface{}) (time.Time, error) {
	if tmpStr, err := v.(json.Number).Int64(); err == nil {
		return time.Unix(tmpStr, 0).UTC(), nil
	}
	return time.Time{}, errors.New("Time should be an unix timestamp")
}


