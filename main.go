package main

import "github.com/sphera-erp/sphera/cli"

//go:generate go run github.com/99designs/gqlgen

func main() {
	cli.Execute()
}
