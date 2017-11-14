#!/bin/bash

# Embed static files with rice
~/go/bin/rice embed-go

# Set correct script dir
DIR="$(cd "$(dirname "$0")" && pwd)"
cd $DIR/..

echo "Building windows binary..."
COMMAND='go build -o build/bins/webby'

~/go/bin/rsrc -manifest build/webby.manifest -ico build/webby.ico
eval "env GOOS=windows GOARCH=386 $COMMAND.exe -ldflags=\"-H windowsgui\""
echo "Windows binary built"
rm -f rsrc.syso

echo "Building linux binary..."
eval "env GOOS=linux GOARCH=amd64 $COMMAND-linux-amd64"
echo "Linux binary built"


