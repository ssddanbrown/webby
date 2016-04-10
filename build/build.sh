#!/bin/bash
#
rice embed-go

# Set correct script dir
DIR="$(cd "$(dirname "$0")" && pwd)"
cd $DIR/..

if [ "$1" != "all" ]
then
	# Create only locally
	go install github.com/ssddanbrown/webby
else
	echo "Building binaries for all platforms"
	COMMAND="go build -o build/bins/webby"

	eval "env GOOS=windows GOARCH=amd64 $COMMAND-windows-amd64.exe"
	echo "Windows binary built"

	eval "env GOOS=linux GOARCH=amd64 $COMMAND-linux-amd64"
	echo "Linux binary built"

	# Create OSX app bundle
	rm -rf build/bins/webby-osx.app
	mkdir build/bins/webby-osx.app
	eval "env GOOS=darwin GOARCH=amd64 $COMMAND-darwin-amd64"
	mv build/bins/webby-darwin-amd64 build/bins/webby-osx.app/webby
	cp build/Info.plist build/bins/webby-osx.app/Info.plist
	echo "OSX binary built"
fi

