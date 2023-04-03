# Copyright 2023 Ben Vogt. All rights reserved.

PWD := $(shell pwd)
OS ?= darwin
ARCH ?= arm64
ENVIRONMENT ?= development

rfind=$(shell find $1 -type f -not -path "./target/*")

target:
	mkdir -p target

output:
	mkdir -p target/output

cloning:
	mkdir -p target/cloning

clean:
	rm -rf target/*

target/gshr-${OS}-${ARCH}-${ENVIRONMENT}.bin:
	go build -o target/gshr-${OS}-${ARCH}-${ENVIRONMENT}.bin main.go

dev: target output cloning target/gshr-${OS}-${ARCH}-${ENVIRONMENT}.bin
	./target/gshr-${OS}-${ARCH}-${ENVIRONMENT}.bin \
    --output=$(PWD)/target/output \
    --clone=$(PWD)/target/cloning && \
    cp styles.css $(PWD)/target/output/ && \
    cd $(PWD)/target/output && \
    python3 -m http.server 8000