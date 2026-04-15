BINARY_NAME=portmap
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

.PHONY: build test lint clean install tidy dev docker-build docker-run docker-stop docker-logs docker-clean

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
	docker build -t $(BINARY_NAME) \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) .

docker-run:
	docker run --rm $(BINARY_NAME)

docker-stop:
	docker compose down

docker-logs:
	docker compose logs -f

docker-clean:
	docker compose down -v
	docker rmi $(BINARY_NAME) 2>/dev/null || true
