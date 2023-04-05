# Copyright 2023 Ben Vogt. All rights reserved.

PWD := $(shell pwd)

rfind=$(shell find $1 -type f -not -path "./target/*")

target:
	mkdir -p target

target/output: target
	mkdir -p target/output

target/cloning: target
	mkdir -p target/cloning

clean:
	rm -rf target/*

target/gshr.bin: Makefile target $(wildcard *.go)
	go build -o target/gshr.bin $(wildcard *.go)

build: Makefile target target/output target/cloning target/gshr.bin
	@# intentionally blank, proxy for prerequisite.

dev: Makefile target target/output target/cloning target/gshr.bin
	./target/gshr.bin \
    --config=$(PWD)/gshr.toml \
    --output=$(PWD)/target/output \
		&& \
    cd $(PWD)/target/output && \
    python3 -m http.server 8000

dev-example-go-git: Makefile target target/output target/cloning target/gshr.bin
	./target/gshr.bin \
    --config=$(PWD)/examples/go-git.toml \
    --output=$(PWD)/target/output \
		&& \
    cd $(PWD)/target/output && \
    python3 -m http.server 8000

dev-example-gshr: Makefile target target/output target/cloning target/gshr.bin
	./target/gshr.bin \
    --config=$(PWD)/examples/ghsr-simple.toml \
    --output=$(PWD)/target/output \
		&& \
    cd $(PWD)/target/output && \
    python3 -m http.server 8000

fmt:
	go fmt