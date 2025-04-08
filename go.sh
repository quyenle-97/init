#!/bin/sh

apk add git

go install github.com/cosmtrek/air@v1.40.4
go mod tidy

air