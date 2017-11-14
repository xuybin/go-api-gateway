#!/bin/bash

set -v on
export CGO_ENABLED=0
if which govendor 2>/dev/null; then
 echo 'govendor exist'
else
 go get -u -v github.com/kardianos/govendor
 echo 'govendor does not exist'
fi


export GOOS=linux GOARCH=amd64
govendor build -o "${PWD##*/}-${GOOS}-${GOARCH}"


export GOOS=windows GOARCH=amd64
govendor build -o "${PWD##*/}-${GOOS}-${GOARCH}.exe"

export GOOS=darwin GOARCH=amd64
go build -o "${PWD##*/}-${GOOS}-${GOARCH}"