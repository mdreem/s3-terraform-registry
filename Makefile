VERSION=$(shell git describe --tags --abbrev=0)
COMMIT=$(shell git rev-list -1 HEAD)

build:
	go build -o bin/main main.go

run:
	go run main.go

test:
	go test -tags testing -v ./... -covermode=count -coverprofile=coverage.out -coverpkg ./...

short_test:
	go test -tags testing -short -v ./... -covermode=count -coverprofile=coverage.out -coverpkg ./...

lint:
	golangci-lint run --config=.github/linters/golangci.yml

clean:
	rm -r bin/**

compile:
	GOOS=linux GOARCH=amd64 go build -ldflags="-X 'main.Version=$(VERSION)' -X 'main.GitCommit=$(COMMIT)'" -o bin/linux-amd64/s3_terraform_registry main.go

docker_build:
	docker build . -f docker/Dockerfile -t ghcr.io/mdreem/terraform-registry:${VERSION}

docker_push:
	docker push ghcr.io/mdreem/terraform-registry:${VERSION}
