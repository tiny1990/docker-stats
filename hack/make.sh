#!/bin/sh

apk add --no-cache git

go get -v -d
go build