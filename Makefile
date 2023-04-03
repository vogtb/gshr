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

dev: target output cloning
	OUTPUT_DIR=$(PWD)/target/output && \
    CLONING_DIR=$(PWD)/target/cloning && \
    go run main.go && \
    cp styles.css OUTPUT_DIR=$(PWD)/target/output/ && \
    cd OUTPUT_DIR=$(PWD)/target/output && \
    python3 -m http.server 8000