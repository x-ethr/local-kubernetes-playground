SUBDIRS := $(wildcard */.)

PWD:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

all :: $(SUBDIRS) deploy

$(SUBDIRS): deploy
	@$(MAKE) -C $@ $(MAKECMDGOALS)

deploy: apply

apply:
	kubectl apply --kustomize .

kustomize:
	@for DIRECTORY in $(shell find . -type d -mindepth 1 -maxdepth 1 -print); do \
		cd "$${DIRECTORY}" ;\
		kubectl apply --wait --kustomize .
		cd "$(PWD)" ;\
	done

clean:
	find . -type d -mindepth 1 -maxdepth 1 -execdir kubectl delete --force --wait --ignore-not-found --all --kustomize . \;

update-server-package:
	@for DIRECTORY in $(shell find . -type d -mindepth 1 -maxdepth 1 -print); do \
		cd "$${DIRECTORY}" ;\
		go get -u github.com/x-ethr/server ;\
		cd "$(PWD)" ;\
	done

reset-versions:
	@for DIRECTORY in $(shell find . -type d -mindepth 1 -maxdepth 1 -print); do \
		cd "$${DIRECTORY}" ;\
		printf "0.0.0" > VERSION ;\
		cd "$(PWD)" ;\
	done
