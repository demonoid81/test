package api

import (
	"fmt"
	"net/http"
)

func succeededHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r)
		w.Header().Set("Server", "A Go Web Server")
		w.WriteHeader(200)
	}
}

func failedHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r)
		w.Header().Set("Server", "A Go Web Server")
		w.WriteHeader(200)
	}
}

func checkHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r)
		w.Header().Set("Server", "A Go Web Server")
		w.WriteHeader(200)
	}
}

func payHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r)
		w.Header().Set("Server", "A Go Web Server")
		w.WriteHeader(200)
	}
}
