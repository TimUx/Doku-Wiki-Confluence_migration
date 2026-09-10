.PHONY: fmt vet test check build release clean

fmt:
	gofmt -w cmd internal

vet:
	go vet ./...

test:
	go test -race -coverprofile=coverage.out ./...

check: vet test

build:
	mkdir -p dist
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o dist/dokuwiki-confluence-migrator ./cmd/migrator

release:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o dist/dokuwiki-confluence-migrator-linux-amd64 ./cmd/migrator
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o dist/dokuwiki-confluence-migrator-linux-arm64 ./cmd/migrator

clean:
	rm -rf dist coverage.out
