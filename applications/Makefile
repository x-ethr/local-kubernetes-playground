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
		if ! [[ "$${DIRECTORY}" =~ ".secrets" ]]; then \
		    cd "$${DIRECTORY}" ;\
		    find . -type d -mindepth 1 -maxdepth 1 -execdir kubectl apply --wait --kustomize . \; ;\
    		cd "$(PWD)" ;\
		fi \
	done

clean:
	find . -type d -mindepth 1 -maxdepth 1 -execdir kubectl delete --force --wait --ignore-not-found --all --kustomize . \;

update-server-package:
	@for DIRECTORY in $(shell find . -type d -mindepth 1 -maxdepth 1 -print); do \
		if ! [[ "$${DIRECTORY}" =~ ".secrets" ]]; then \
		    cd "$${DIRECTORY}" ;\
		    go get -u github.com/x-ethr/server ;\
    		cd "$(PWD)" ;\
		fi \
	done
