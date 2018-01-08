CURRENT_DIR = $(shell pwd)
PROJECT_ROOT ?= $(shell dirname $(CURRENT_DIR))

test:
	@ echo "PROJECT_ROOT=$(PROJECT_ROOT)"
	@ echo "CURRENT_DIR=$(CURRENT_DIR)"

deploy: vendor
	@# To deploy for GAE, use ./gopath instead of default GOPATH.
	# TODO go-chat itself is not contained in GOPATH or gopath.
	## GOPATH=$(CURRENT_DIR)/gopath gcloud app deploy $(CURRENT_DIR)/app/app.yaml

vendor: $(PROJECT_ROOT)/Gopkg.lock
	ln -sf $(PROJECT_ROOT)/vendor $(CURRENT_DIR)/gopath/src
