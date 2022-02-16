#!/usr/bin/env sh
for os in "linux"
do
    for arch in "amd64" "arm64"
    do
        CGO_ENABLED=0 \
        GOOS=${os} \
        GOARCH=${arch} \
        go build -o ./bin/dbt-pg-proxy-${os}-${arch} cmd/main.go
    done
done
