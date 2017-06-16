TAG ?= latest
REPO ?= foomo/petze

docker-build:
	GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -o docker/petze petze.go
	docker build -t $(REPO):$(TAG) docker
	rm -vf docker/petze

docker-push:
	docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD";
	docker push $(REPO)
