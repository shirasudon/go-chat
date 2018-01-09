CURRENT_DIR = $(shell pwd)
PROJECT_ROOT ?= $(shell dirname $(CURRENT_DIR))

CURRENT_GOPATH_ROOT = $(CURRENT_DIR)/_gopath

PROJECT_LIB_NAMES = server chat domain infra ws 


.PHONY: help
help:
	@echo "test   # run local server"
	@echo "deploy # deploy the application for GAE"
	@echo "vendor # collect dependencies for deploy"
	@echo "clean-vendor # clean dependencies for deploy"

.PHONY: test
test:
	dev_appserver.py app/app.yaml

.PHONY: deploy
deploy: vendor
	@# To deploy for GAE, use ./gopath instead of default GOPATH.
	GOPATH=$(CURRENT_GOPATH_ROOT)/vendor:$(CURRENT_GOPATH_ROOT)  gcloud app deploy $(CURRENT_DIR)/app/app.yaml

.PHONY: vendor
vendor: $(PROJECT_ROOT)/Gopkg.lock clean-vendor
	mkdir -p $(CURRENT_GOPATH_ROOT)/src/github.com/shirasudon/go-chat
	mkdir -p $(CURRENT_GOPATH_ROOT)/vendor
	ln -sf $(PROJECT_ROOT)/vendor $(CURRENT_GOPATH_ROOT)/vendor/src
	@# go-chat sub packages are placed under ./src/
	@for lib in $(PROJECT_LIB_NAMES); do\
		echo "ln -sf $(PROJECT_ROOT)/$$lib $(CURRENT_GOPATH_ROOT)/src/github.com/shirasudon/go-chat/$$lib";\
		ln -sf $(PROJECT_ROOT)/$$lib $(CURRENT_GOPATH_ROOT)/src/github.com/shirasudon/go-chat/$$lib || exit 50;\
	done 

.PHONY: clean-vendor
clean-vendor: 
	rm -f $(CURRENT_GOPATH_ROOT)/vendor/src
	@for lib in $(PROJECT_LIB_NAMES); do\
		echo "rm -f $(CURRENT_GOPATH_ROOT)/src/github.com/shirasudon/go-chat/$$lib";\
		rm -f $(CURRENT_GOPATH_ROOT)/src/github.com/shirasudon/go-chat/$$lib ;\
	done 
