#!/bin/bash

rice embed-go

# Set correct script dir
DIR="$(cd "$(dirname "$0")" && pwd)"
cd $DIR/..

if [ "$1" != "all" ]
then
	# Create only locally
	go install github.com/ssddanbrown/webby
	echo "Webby installed"
else
	echo "Building binaries for all platforms"
	COMMAND="go build -o build/bins/webby"

	rsrc -manifest build/webby.manifest -ico build/webby.ico,build/webby-64.ico,build/webby-32.ico,build/webby-16.ico
	eval "env GOOS=windows GOARCH=386 $COMMAND-windows-386.exe"
	echo "Windows binary built"
	rm -f rsrc.syso

	eval "env GOOS=linux GOARCH=amd64 $COMMAND-linux-amd64"
	echo "Linux binary built"

	# Create OSX app bundle
	rm -rf build/bins/webby-osx.app
	mkdir build/bins/webby-osx.app
	eval "env GOOS=darwin GOARCH=amd64 $COMMAND-darwin-amd64"
	mv build/bins/webby-darwin-amd64 build/bins/webby-osx.app/webby
	cp build/mac-launch-wrapper.sh build/bins/webby-osx.app/wrapper
	cp build/webby.icns build/bins/webby-osx.app/webby.icns
	cp build/Info.plist build/bins/webby-osx.app/Info.plist
	echo "OSX binary built"
fi

