REMOTE_USER ?= $(shell whoami)
INSTANCE ?= $(shell cat instance_name)
GO_PACKAGES := $(shell go list ./...)

.PHONY: build
build: fmt vet
	go build ./cmd/bot

.PHONY: run
run:
	go run ./cmd/bot

.PHONY: deploy
deploy:
	scp ./bot $(INSTANCE):/home/$(REMOTE_USER)/ind/bot

.PHONY: deploy_all
deploy_all: deploy
	scp ./scripts/start.sh ./scripts/stop.sh $(INSTANCE):/home/$(REMOTE_USER)/ind/

.PHONY: fmt
fmt:
	@go fmt $(GO_PACKAGES)

.PHONY: vet
vet:
	@go vet $(GO_PACKAGES)

.PHONY: docker
docker:
	@docker build -t indbot .
