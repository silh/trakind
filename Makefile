INSTANCE ?= $(shell cat instance_name)

.PHONY: build
build:
	go build ./cmd/bot

.PHONY: run
run:
	go run ./cmd/bot

.PHONY: deploy
deploy:
	scp ./bot ec2-user@$(INSTANCE):/home/ec2-user/bot

.PHONY: deploy_all
deploy_all: deploy
	scp ./scripts/ind_start.sh ./scripts/ind_stop.sh  ec2-user@$(INSTANCE):/home/ec2-user/
