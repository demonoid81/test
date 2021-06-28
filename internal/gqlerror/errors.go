package gqlerror

import "fmt"

type GQLError struct {
	kind      fmt.Stringer
	status    Status
	arguments map[string]string
	details   string
	message   string
}

type Status int
