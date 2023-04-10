# Copyright 2023 Ben Vogt. All rights reserved.

target:
	mkdir -p target

target/output: target
	mkdir -p target/output

clean:
	rm -rf target

deps:
	go mod download

fmt:
	go fmt

install:
	go install

target/gshr.bin: Makefile target $(wildcard *.go)
	go build -o target/gshr.bin $(wildcard *.go)

build: Makefile target target/output  target/gshr.bin
	@# intentionally blank, proxy for prerequisite.

dev: Makefile target target/output target/gshr.bin
	./target/gshr.bin -c=dev-config-gshr.toml -o=target/output && \
    cd target/output && \
    python3 -m http.server 80

dev-example-go-git: Makefile target target/output target/gshr.bin
	./target/gshr.bin -c=example-config-gshr-simple.toml -o=target/output && \
    cd target/output && \
    python3 -m http.server 80

dev-example-gshr-simple: Makefile target target/output target/gshr.bin
	./target/gshr.bin -c=example-config-gshr-simple.toml -o=target/output && \
    cd target/output && \
    python3 -m http.server 80
