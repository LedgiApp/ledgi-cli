BINARY := ledgi
DIST := dist
MODULE := github.com/LedgiApp/ledgi-cli

.PHONY: build build-all test clean

build:
	go build -o $(DIST)/$(BINARY) .

build-all:
	GOOS=linux  GOARCH=amd64 go build -o $(DIST)/$(BINARY)-linux-amd64 .
	GOOS=linux  GOARCH=arm64 go build -o $(DIST)/$(BINARY)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 go build -o $(DIST)/$(BINARY)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o $(DIST)/$(BINARY)-darwin-arm64 .

test:
	go test ./...

clean:
	rm -rf $(DIST)
