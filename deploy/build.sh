#!/bin/bash
set -e
cd "$(dirname "$0")/.."

ARCH=${1:-amd64}
OUTPUT="dist/sub-linux-${ARCH}.tar.gz"

echo "Building sub for ${ARCH}..."
GOARCH=amd64 go build -ldflags="-s -w" -o dist/sub .

mkdir -p dist
tar -czf "${OUTPUT}" \
    dist/sub \
    pupmsub.service \
    install.sh \
    pupmsub.sh \
    README.md

echo "Built: ${OUTPUT}"
ls -lh "${OUTPUT}"
