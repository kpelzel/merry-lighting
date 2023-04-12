# Merry Lighting
Merry lighting receives incoming sACN packets (from qlc+ for example) and sends out bluetooth packets for "happy lighting" branded lights

Thanks to https://github.com/madhead/saberlight/blob/master/protocols/Triones/protocol.md for the BLE protocol

## Cross compile for windows from mac
```
brew install mingw-w64
env GOOS="windows" GOARCH="amd64" CGO_ENABLED="1" CC="x86_64-w64-mingw32-gcc" go build
```