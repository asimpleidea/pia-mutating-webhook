# Image URL to use all building/pushing image targets.
IMG ?= {CONTAINER_IMAGE}

# Run tests.
test: fmt vet
	go test -coverprofile cover.out

# Build the binary.
build: fmt vet
	go build -o bin/users-api main.go

# Run.
run: fmt vet
	go run ./main.go

# Run go fmt against code.
fmt:
	go fmt ./...

# Run go vet against code.
vet:
	go vet ./...

# Build the docker image.
docker-build: test
	docker build . -t ${IMG}

# Push the docker image.
docker-push:
	docker push ${IMG}