all: build

build:
	go build -o bin/storageX ./cmd

run:
	./bin/storageX --config config/config.yaml

clean:
	rm -rf bin/
