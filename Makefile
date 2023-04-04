# Copyright 2023 Ben Vogt. All rights reserved.

PWD := $(shell pwd)
OS ?= darwin
ARCH ?= arm64
ENVIRONMENT ?= development

rfind=$(shell find $1 -type f -not -path "./target/*")

target:
	mkdir -p target

target/output: target
	mkdir -p target/output

target/cloning: target
	mkdir -p target/cloning

clean:
	rm -rf target/*

target/gshr-${OS}-${ARCH}-${ENVIRONMENT}.bin: Makefile target
	go build -o target/gshr-${OS}-${ARCH}-${ENVIRONMENT}.bin $(wildcard *.go)

dev: Makefile target target/output target/cloning target/gshr-${OS}-${ARCH}-${ENVIRONMENT}.bin
	./target/gshr-${OS}-${ARCH}-${ENVIRONMENT}.bin \
    --name="gshr" \
    --desc="git static host repo -- generates static html for repo" \
    --repo=/Users/bvogt/dev/src/ben/gshr \
    --output=$(PWD)/target/output \
    --clone=$(PWD)/target/cloning \
		&& \
    cp gshr.css $(PWD)/target/output/ && \
    cd $(PWD)/target/output && \
    python3 -m http.server 8000

fmt:
	go fmt