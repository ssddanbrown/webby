#!/bin/bash
#
rice embed-go

if [ "$1" != "all" ]
then
	# Create only locally
	go install github.com/ssddanbrown/webby
else
	echo "Building binaries for all platforms"
	COMMAND="go build -o bins/webby"

	eval "env GOOS=windows GOARCH=amd64 $COMMAND-windows-amd64.exe"
	echo "Windows binary built"

	eval "env GOOS=linux GOARCH=amd64 $COMMAND-linux-amd64"
	echo "Linux binary built"

	eval "env GOOS=darwin GOARCH=amd64 $COMMAND-darwin-amd64"
	echo "OSX binary built"
fi

