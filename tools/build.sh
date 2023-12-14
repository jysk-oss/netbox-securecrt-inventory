mkdir -p dist/darwin/arm64
mkdir -p dist/darwin/amd64

cp -r tools/assets/securecrt-inventory.app dist/darwin/arm64
cp -r tools/assets/securecrt-inventory.app dist/darwin/amd64

CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -ldflags "-H=windowsgui" -o dist/securecrt-inventory.exe main.go
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64  go build -o dist/darwin/arm64/securecrt-inventory.app/Contents/MacOS/securecrt-inventory main.go
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64  go build -o dist/darwin/amd64/securecrt-inventory.app/Contents/MacOS/securecrt-inventory main.go

zip -r dist/securecrt-inventory-darwin-arm64.zip dist/darwin/arm64/securecrt-inventory.app 
zip -r dist/securecrt-inventory-darwin-amd64.zip dist/darwin/amd64/securecrt-inventory.app 
zip -r dist/securecrt-inventory-windows-amd64.zip dist/securecrt-inventory.exe

rm dist/securecrt-inventory.exe
rm -rf dist/darwin/arm64
rm -rf dist/darwin/amd64
