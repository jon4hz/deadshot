#!/bin/sh
set -e
rm -rf manpages
mkdir manpages
go run ./cmd/deadshot/main.go man | gzip -c > manpages/deadshot.1.gz