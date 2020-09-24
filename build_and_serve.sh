cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" public/
GOOS=js GOARCH=wasm go build -o public/game.wasm ./wasm
go run ./server
