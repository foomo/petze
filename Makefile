TAG ?= latest

docker-build:
	GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -o docker/petze petze.go
	docker build -t foomo/petze:$(TAG) docker
	rm -vf docker/petze