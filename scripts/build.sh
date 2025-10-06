#!/bin/bash

# Build script for envswitch with version information from git tags

set -e

# Get version information from git
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(date -u '+%Y-%m-%d_%H:%M:%S')

# Build output directory
OUTPUT_DIR=${OUTPUT_DIR:-"./bin"}
mkdir -p "$OUTPUT_DIR"

# Binary name
BINARY_NAME=${BINARY_NAME:-"envswitch"}

# Build flags
LDFLAGS="-X github.com/hugofrely/envswitch/cmd.Version=${VERSION}"
LDFLAGS="${LDFLAGS} -X github.com/hugofrely/envswitch/cmd.GitCommit=${COMMIT}"
LDFLAGS="${LDFLAGS} -X github.com/hugofrely/envswitch/cmd.BuildDate=${DATE}"

# Add strip flags for smaller binary in production
if [ "$PRODUCTION" = "true" ]; then
    LDFLAGS="${LDFLAGS} -s -w"
fi

echo "Building ${BINARY_NAME}..."
echo "  Version:    ${VERSION}"
echo "  Commit:     ${COMMIT}"
echo "  Build Date: ${DATE}"
echo ""

# Build for current platform
go build -ldflags="${LDFLAGS}" -o "${OUTPUT_DIR}/${BINARY_NAME}" .

echo "✅ Build complete: ${OUTPUT_DIR}/${BINARY_NAME}"
echo ""
echo "Run with: ${OUTPUT_DIR}/${BINARY_NAME} --version"
