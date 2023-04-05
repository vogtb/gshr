# Copyright 2023 Ben Vogt. All rights reserved.

PWD := $(shell pwd)

rfind=$(shell find $1 -type f -not -path "./target/*")

target:
	mkdir -p target

target/output: target
	mkdir -p target/output

clean:
	rm -rf target/*

target/gshr.bin: Makefile target $(wildcard *.go)
	go build -o target/gshr.bin $(wildcard *.go)

build: Makefile target target/output  target/gshr.bin
	@# intentionally blank, proxy for prerequisite.

dev: Makefile target target/output  target/gshr.bin
	./target/gshr.bin -c=gshr.toml -o=target/output \
    cd target/output && python3 -m http.server 80

dev-example-go-git: Makefile target target/output  target/gshr.bin
	./target/gshr.bin -c=examples/go-git.toml -o=target/output \
    cd target/output && python3 -m http.server 80

dev-example-gshr: Makefile target target/output  target/gshr.bin
	./target/gshr.bin -c=examples/gshr-simple.toml -o=target/output \
    cd target/output && python3 -m http.server 80

fmt:
	go fmt