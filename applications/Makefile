SUBDIRS := $(wildcard */.)

PWD:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

all:
	@for DIRECTORY in $(shell find . -type d -mindepth 1 -maxdepth 1 -print); do \
		cd "$${DIRECTORY}" ;\
		make ;\
		cd "$(PWD)" ;\
	done

update:
	@for DIRECTORY in $(shell find . -type d -mindepth 1 -maxdepth 1 -print); do \
		cd "$${DIRECTORY}" ;\
		go get -u ./... ;\
		go mod vendor ;\
		cd "$(PWD)" ;\
	done

kustomize:
	@for DIRECTORY in $(shell find . -type d -mindepth 1 -maxdepth 1 -print); do \
		cd "$${DIRECTORY}" ;\
		kubectl apply --wait --kustomize . ;\
		cd "$(PWD)" ;\
	done

clean:
	@for DIRECTORY in $(shell find . -type d -mindepth 1 -maxdepth 1 -print); do \
		cd "$${DIRECTORY}" ;\
		kubectl delete --wait --ignore-not-found --all --kustomize . ;\
		cd "$(PWD)" ;\
	done

vet:
	@for DIRECTORY in $(shell find . -type d -mindepth 1 -maxdepth 1 -print); do \
		cd "$${DIRECTORY}" ;\
		go vet ./... ;\
		cd "$(PWD)" ;\
	done

reset-versions:
	@for DIRECTORY in $(shell find . -type d -mindepth 1 -maxdepth 1 -print); do \
		cd "$${DIRECTORY}" ;\
		printf "0.0.0" > VERSION ;\
		cd "$(PWD)" ;\
	done
