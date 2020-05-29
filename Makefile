.PHONY: all
all: bin/astats

dist:
	mkdir -p bin

bin:
	mkdir -p bin

bin/astats: $(shell find . -name '*.go') go.mod bin
	cd cmd/astats && go build -o ../../$@

.PHONY: linux
linux: dist
	docker run --rm -i -t -v $(PWD):/src -w /src/cmd/astats golang:1.14 go build -o ../../dist/astats

.PHONY: test
test:
	go test ./... -v

.PHONY: clean
clean:
	rm -rf bin dist
