VERSION 					?=v1.0.4
PG_VERSION 					?=15

.PHONY: docker
docker:
	@docker build -t skydev/db-backup:$(VERSION)-pg$(PG_VERSION) --build-arg PG_VERSION=$(PG_VERSION) -f ci/Dockerfile .
	@docker push skydev/db-backup:$(VERSION)-pg$(PG_VERSION)

.PHONY: lint
lint:
	golangci-lint run