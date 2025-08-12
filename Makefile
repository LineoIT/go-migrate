
build:
	mkdir -p build
	GOOS=linux GOARCH=amd64 go build -o build/migrate-linux cli/*.go
	GOOS=darwin GOARCH=amd64 go build -o build/migrate-macos cli/*.go
	GOOS=darwin GOARCH=arm64 go build -o build/migrate-macos-arm64 cli/*.go
	GOOS=windows GOARCH=amd64 go build -o build/migrate-windows.exe cli/*.go
	echo "Builds completed."
