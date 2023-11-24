.PHONY: clean compile build upload push all
OUTPUT_DIR=./out
DOCKER_IMG=patcharp/lession1
VERSION=$(shell git describe --always --dirty)
DOCKER_TAG=$(DOCKER_IMG)

# CI automate version set
ifeq ($(CI), true)
	VERSION=$(CI_BUILD_REF_NAME)-build$(CI_PIPELINE_ID)
endif

ifeq ($(VERSION), "")
	DOCKER_TAG=$(DOCKER_IMG):latest
endif

clean:
	rm -rf $(OUTPUT_DIR)

download:
	go mod download

run:
	go run -ldflags "-X 'main.Version=$(VERSION)'" main.go

build: clean
	go build -ldflags "-X 'main.Version=$(VERSION)'" -o $(OUTPUT_DIR)/app main.go

docker:
	docker build -t $(DOCKER_TAG) .
