#! /usr/bin/env bash

# Architectures we want to target
archs=(amd64 arm arm64)

echo "Starting Compilation"

echo "Compiling macOS binaries"
go build -o mdconverter_mac-amd64
echo "Compiled binary from mac/amd64"
env GOOS=darwin GOARCH=arm64 go build -o mdconverter_mac-arm64
echo "Compiled binary from mac/arm64"

echo "Compiling Linux binaries"
for arch in ${archs[@]}
do
	env GOOS=linux GOARCH=${arch} go build -o mdconverter_lin_${arch}
  echo "Compiled linux/${arch}"
done

echo "Compiling Windows binaries"
for arch in ${archs[@]}
do
	env GOOS=windows GOARCH=${arch} go build -o mdconverter_win_${arch}
  echo "Compiled windows/${arch}"
done