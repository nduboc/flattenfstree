#!/bin/bash

GOOS=darwin  GOARCH=amd64  go build -o ./flattenfstree.macos
GOOS=windows GOARCH=amd64  go build -o ./flattenfstree.exe
#GOOS=linux   GOARCH=amd64  go build -o ./flattenfstree.linux
