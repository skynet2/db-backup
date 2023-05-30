.PHONY: docker
docker:
	@docker build -t skydev/db-backup:0.0.0.1-pg14 -f ci/Dockerfile-postgres14 .