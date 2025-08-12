
build:
	mkdir -p build
	GOOS=linux GOARCH=amd64 go build -o build/migrate-linux *.go
	GOOS=darwin GOARCH=amd64 go build -o build/migrate-macos *.go
	GOOS=darwin GOARCH=arm64 go build -o build/migrate-macos-arm64 *.go
	GOOS=windows GOARCH=amd64 go build -o build/migrate-windows.exe *.go
	echo "Builds completed."
