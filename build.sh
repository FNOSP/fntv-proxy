#!/bin/sh

set -e

OUTPUT_BASE="build"
PROJECT_NAME="fntv-proxy"

# Clean up previous build
echo "Cleaning up $OUTPUT_BASE directory..."
rm -rf "$OUTPUT_BASE"
mkdir -p "$OUTPUT_BASE"

# Define targets as a space-separated string for sh compatibility
TARGETS="windows/amd64 windows/arm64 linux/amd64 linux/arm64 darwin/amd64 darwin/arm64"

for target in $TARGETS; do
    # Split target into OS and ARCH using string manipulation
    OS=${target%/*}
    ARCH=${target#*/}
    
    FOLDER_ARCH="$ARCH"
    if [ "$ARCH" = "arm64" ]; then
        FOLDER_ARCH="aarch64"
    fi

    OUTPUT_DIR="$OUTPUT_BASE/${OS}_${FOLDER_ARCH}"
    mkdir -p "$OUTPUT_DIR"

    EXE_NAME="$PROJECT_NAME"
    if [ "$OS" = "windows" ]; then
        EXE_NAME="${PROJECT_NAME}.exe"
    fi

    echo "Building for $OS/$ARCH -> $OUTPUT_DIR..."

    export CGO_ENABLED=0
    export GOOS="$OS"
    export GOARCH="$ARCH"

    go build -trimpath -ldflags "-s -w" -o "$OUTPUT_DIR/$EXE_NAME" .
done

echo "Build complete. Artifacts are in $OUTPUT_BASE/"
