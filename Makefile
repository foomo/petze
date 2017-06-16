TAG ?= latest
REPO ?= foomo/petze

build-docker:
	GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -o docker/petze petze.go
	docker build -t $(REPO):$(TAG) docker
	rm -vf docker/petze

push-docker:
	docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD";
	docker push $(REPO)
