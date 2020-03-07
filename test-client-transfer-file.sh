#!/bin/sh
go run github.com/hellgate75/go-tcp-client -verbosity DEBUG transfer-file "./main.go" "$(echo $HOME)/tmpGoServer/main-remote.go"
ls -latr $(echo $HOME)/tmpGoServer
rm -Rf $(echo $HOME)/tmpGoServer
