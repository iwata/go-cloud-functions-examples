GOFILES=$(shell find . -name '*.go')
GOPACKAGES=$(shell go list ./...)
COVER_DIR='./cover'
BINFILE='bin/backup'

default: build

build: $(GOFILES)
	go build -o ${BINFILE} -v -race ./cmd/firestorebackup

test: $(GOFILES)
	go test -v -race ./...

cov: $(GOFILES)
	go test -coverprofile=${COVER_DIR}/cover.out -race ./...
	go tool cover -html=${COVER_DIR}/cover.out -o ${COVER_DIR}/cover.html

lint: $(GOFILES)
	@golangci-lint run ./...

.PHONY: clean
clean:
	go clean
	rm -f ${BINFILE}

.PHONY: watch
watch:
	reflex -r '\.go$$' -- sh -c 'make lint; make test'
