#!/bin/sh
set -e
rm -rf completions
mkdir completions
for sh in bash zsh fish; do
	go run ./cmd/deadshot/main.go completion "$sh" >"completions/deadshot.$sh"
done