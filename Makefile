# Get version from git hash
current_tag := $(shell git tag --points-at HEAD)
git_hash := $(shell git tag --points-at HEAD && [ ! -z ${current_tag} ] || git rev-parse --short HEAD || echo 'UNKNOWN')

# Get current date
current_time = $(shell date +"%Y-%m-%d:T%H:%M:%S")

# Add linker flags
linker_flags = '-s -X main.buildTime=${current_time} -X main.version=${git_hash}'

BINARY_NAME=tfpin
GO_BINARY=$$(which go)

# Build binaries for current OS and Linux
.PHONY:
build:
	$(GO_BINARY) build -ldflags=${linker_flags} -o=./bin/$(BINARY_NAME) main.go
	GOOS=linux GOARCH=amd64 $(GO_BINARY) build -ldflags=${linker_flags} -o=./bin/linux_amd64/$(BINARY_NAME) main.go

run:
	./bin/${BINARY_NAME}

build_run: build run

clean:
	$(GO_BINARY) clean
	rm -rf ./bin
