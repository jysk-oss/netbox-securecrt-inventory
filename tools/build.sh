#!/usr/bin/env bash
set -e

mkdir -p dist

case "$(uname)" in
  Darwin)
    mkdir -p dist/darwin/{arm64,amd64}

    cp -r tools/assets/securecrt-inventory.app dist/darwin/arm64
    cp -r tools/assets/securecrt-inventory.app dist/darwin/amd64

    CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 \
      go build -o dist/darwin/arm64/securecrt-inventory.app/Contents/MacOS/securecrt-inventory main.go

    CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 \
      go build -o dist/darwin/amd64/securecrt-inventory.app/Contents/MacOS/securecrt-inventory main.go

    (cd dist/darwin/arm64 && zip -r ../../securecrt-inventory-darwin-arm64.zip securecrt-inventory.app)
    (cd dist/darwin/amd64 && zip -r ../../securecrt-inventory-darwin-amd64.zip securecrt-inventory.app)
    ;;

  Linux)
    mkdir -p dist/linux/amd64

    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
      go build -o dist/linux/amd64/securecrt-inventory main.go

    (cd dist/linux/amd64 && zip -r ../../securecrt-inventory-linux-amd64.zip securecrt-inventory)
    ;;

  MINGW*|MSYS*|CYGWIN*)
    CGO_ENABLED=1 GOOS=windows GOARCH=amd64 \
      go build -ldflags "-H=windowsgui" -o dist/securecrt-inventory.exe main.go

    (cd dist && zip -r securecrt-inventory-windows-amd64.zip securecrt-inventory.exe)
    ;;
esac