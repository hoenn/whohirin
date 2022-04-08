.PHONY: build
build:
	go build -o bin/whohirin .

binaries:
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o bin/whohirin-amd64-darwin .
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/whohirin-amd64-linux .

clean:
	rm ./bin/*
