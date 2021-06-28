package models

import (
	"encoding/json"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/pkg/errors"
	"io"
)

// Creates a marshaller which converts a time.Time (date) to a string
func MarshalJSON(j map[string]interface{}) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		bytes, err := json.Marshal(j)
		if err != nil {
			fmt.Print(errors.Wrapf(err, "while marshalling %+v scalar object", j))
			return
		}
		_, err = w.Write(bytes)
		if err != nil {
			fmt.Print(errors.Wrapf(err, "while writing marshalled %+v object", j))
			return
		}
	})
}

// Unmarshalls a string to a time.Time (date)
func UnmarshalJSON(v interface{}) (map[string]interface{}, error) {
	if in, ok := v.(string); ok {
		var jsonMap map[string]interface{}
		err := json.Unmarshal([]byte(in), &jsonMap)
		if err != nil {
			return nil, errors.Wrapf(err, "while unmarshalling %+v scalar object", v)
		}
		v = jsonMap
	}

	value, ok := v.(map[string]interface{})
	if !ok {
		return nil ,errors.Errorf("Unable to convert interface %T to map[string]interface{}", v)
	}
	return value, nil
}
