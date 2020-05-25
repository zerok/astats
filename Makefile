all: bin/astats

bin:
	mkdir -p bin

bin/astats: $(shell find . -name '*.go') go.mod bin
	cd cmd/astats && go build -o ../../$@

test:
	go test ./... -v

clean:
	rm -rf bin
