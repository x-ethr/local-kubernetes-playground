SUBDIRS := $(wildcard */.)

all :: $(SUBDIRS) deploy

$(SUBDIRS): deploy
	@$(MAKE) -C $@ $(MAKECMDGOALS)

deploy: apply

apply:
	kubectl apply --kustomize .

kustomize:
	find . -type d -mindepth 1 -maxdepth 1 -execdir kubectl apply --wait --kustomize . \;

clean:
	find . -type d -mindepth 1 -maxdepth 1 -execdir kubectl delete --force --wait --ignore-not-found --all --kustomize . \;

update-server-package:
	find . -type d -mindepth 1 -maxdepth 1 -execdir go get -u github.com/x-ethr/server \;
