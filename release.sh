#!/bin/bash

for os in darwin windows linux; do
    for arch in 386; do
        env GOOS="$os" GOARCH="$arch" go build -o "./.bin/thyme-$os-$arch" ./cmd/thyme
    done
done
