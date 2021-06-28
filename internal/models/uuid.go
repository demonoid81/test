package models

import (
	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"io"
	"log"
	"strconv"
)

func MarshalUUID(u uuid.UUID) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, err := io.WriteString(w, strconv.Quote(u.String()))
		if err != nil {
			log.Print(errors.Wrap(err, "while writing marshalled timestamp"))
			return
		}
	})
}

func UnmarshalUUID(v interface{}) (uuid.UUID, error) {
	if tmpStr, ok := v.(string); ok {
		return uuid.Parse(tmpStr)
	}
	return uuid.Nil, errors.New("Time should be an unix timestamp")
}
