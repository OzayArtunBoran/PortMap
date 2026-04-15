BINARY_NAME=portmap
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

.PHONY: build test lint clean install tidy dev

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) .

test:
	go test ./... -v -count=1

lint:
	go vet ./...

clean:
	rm -rf bin/

tidy:
	go mod tidy

install: build
	cp bin/$(BINARY_NAME) $(GOPATH)/bin/

dev: build
	./bin/$(BINARY_NAME)

# Docker
docker-build:
	docker build -t $(BINARY_NAME) .

docker-run:
	docker run --rm $(BINARY_NAME)
