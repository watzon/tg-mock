.PHONY: build run generate test clean docker-build docker-run

# Docker settings
DOCKER_IMAGE ?= ghcr.io/watzon/tg-mock
DOCKER_TAG ?= latest

build:
	go build -o bin/tg-mock ./cmd/tg-mock
	go build -o bin/codegen ./cmd/codegen

run: build
	./bin/tg-mock

generate:
	go run ./cmd/codegen -spec spec/api.json -out gen

test:
	go test -v ./...

clean:
	rm -rf bin/

fetch-spec:
	curl -sL https://raw.githubusercontent.com/PaulSonOfLars/telegram-bot-api-spec/main/api.json > spec/api.json

fetch-errors:
	curl -sL https://raw.githubusercontent.com/TelegramBotAPI/errors/master/errors.json > errors/errors.json

# Docker targets
docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: docker-build
	docker run --rm -p 8081:8081 $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-push: docker-build
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
