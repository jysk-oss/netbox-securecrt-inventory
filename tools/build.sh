mkdir -p dist/darwin/arm64
mkdir -p dist/darwin/amd64
mkdir -p dist/linux/amd64

cp -r tools/assets/securecrt-inventory.app dist/darwin/arm64
cp -r tools/assets/securecrt-inventory.app dist/darwin/amd64

CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -ldflags "-H=windowsgui" -o dist/securecrt-inventory.exe main.go
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o dist/darwin/arm64/securecrt-inventory.app/Contents/MacOS/securecrt-inventory main.go
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o dist/darwin/amd64/securecrt-inventory.app/Contents/MacOS/securecrt-inventory main.go
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o dist/linux/amd64/securecrt-inventory main.go

cd dist
cd darwin/arm64 && zip -r ../../securecrt-inventory-darwin-arm64.zip securecrt-inventory.app
cd ../amd64 && zip -r ../../securecrt-inventory-darwin-amd64.zip securecrt-inventory.app
cd ../../linux/amd64 && zip -r ../../securecrt-inventory-linux-amd64.zip securecrt-inventory
cd ../../ && zip -r -j securecrt-inventory-windows-amd64.zip securecrt-inventory.exe

rm securecrt-inventory.exe
rm -rf darwin
rm -rf linux
