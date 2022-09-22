BINARY_NAME=tfpin
GO_BINARY=$$(which go)

build:
	$(GO_BINARY) build -o ${BINARY_NAME} main.go

run:
	./${BINARY_NAME}

build_run: build run

clean:
	$(GO_BINARY) clean
	rm ${BINARY_NAME}
