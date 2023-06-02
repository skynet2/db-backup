VERSION 					?=v1.0.0

.PHONY: docker
docker:
	@docker build -t skydev/db-backup:$(VERSION)-pg14 -f ci/Dockerfile-postgres14 .
	@docker push -t skydev/db-backup:$(VERSION)-pg14