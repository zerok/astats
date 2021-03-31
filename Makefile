.PHONY: all
all: bin/astats

dist:
	mkdir -p bin

bin:
	mkdir -p bin

bin/astats: $(shell find . -name '*.go') go.mod bin
	cd cmd/astats && go build -mod=mod -o ../../$@

.PHONY: linux
linux: dist
	docker run --rm -i -t -v $(PWD):/src -w /src/cmd/astats golang:1.16 go build -mod=mod -o ../../dist/astats

.PHONY: test
test:
	go test -mod=mod ./... -v

.PHONY: clean
clean:
	rm -rf bin dist
