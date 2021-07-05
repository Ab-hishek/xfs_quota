# Makefile for building lvm-exporter
# Reference Guide - https://www.gnu.org/software/make/manual/make.html

#
# Internal variables or constants.
# NOTE - These will be executed when any make target is invoked.
#
IS_DOCKER_INSTALLED = $(shell which docker >> /dev/null 2>&1; echo $$?)

#docker info
ACCOUNT ?= abhishek09dh
XFSQUOTAREPO ?= xfs_quota
IMGTAG ?= v1.7

.PHONY: all
all: deps go-deps fmt xfs-quota

.PHONY: help
help:
	@echo ""
	@echo "Usage:-"
	@echo "\tmake all   -- [default] runs all checks, builds and push the Kubera Portal backend image"
	@echo ""

.PHONY: deps

deps:
	@echo "------------------"
	@echo "--> Check the Docker deps"
	@echo "------------------"
	@if [ $(IS_DOCKER_INSTALLED) -eq 1 ]; \
		then echo "" \
		&& echo "ERROR:\tdocker is not installed. Please install it before build." \
		&& echo "" \
		&& exit 1; \
		fi;

# Run go fmt against code
.PHONY: fmt
fmt:
	@go fmt ./...

.PHONY: go-deps
go-deps:
	@echo "--> Tidying up submodules"
	@go mod tidy
	@echo "--> Verifying submodules"
	@go mod verify

.PHONY: vendor
vendor: go.mod go.sum go-deps
	@go mod vendor

.PHONY: xfs-quota

xfs-quota:
	@echo "------------------"
	@echo "--> Build lvm-exporter image"
	@echo "------------------"
	docker build -t $(ACCOUNT)/$(XFSQUOTAREPO):$(IMGTAG) -f Dockerfile .

push:
	@echo "------------------"
	@echo "--> Pushing lvm-exporter image"
	@echo "------------------"
	docker push ${ACCOUNT}/${XFSQUOTAREPO}:${IMGTAG}