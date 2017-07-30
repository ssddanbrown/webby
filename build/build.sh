#!/bin/bash

# Embed static files with rice
rice embed-go

# Set correct script dir
DIR="$(cd "$(dirname "$0")" && pwd)"
cd $DIR/..

echo "Building windows binary"
COMMAND='go build -ldflags="-H windowsgui" -o build/bins/webby.exe'

rsrc -manifest build/webby.manifest -ico build/webby.ico
eval "env GOOS=windows GOARCH=386 $COMMAND"
echo "Windows binary built"
rm -f rsrc.syso


