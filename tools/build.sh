mkdir -p dist
cp -r tools/assets/securecrt-inventory.app dist

CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -ldflags "-H=windowsgui" -o dist/securecrt-inventory.exe main.go
CGO_ENABLED=1 go build -o dist/securecrt-inventory.app/Contents/MacOS/securecrt-inventory main.go