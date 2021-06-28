#!/bin/bash
reflex -r '\.go' -s -- sh -c "clear && go run main.go server --tarantoolUrl=http://localhost:8787"
