default: build

build:
	go build -v ./...

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

# See https://golangci-lint.run/
lint:
	golangci-lint run -c .golangci.toml ./...

generate:
	go generate ./...

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -coverprofile=cover.out -parallel=4 ./...
	go tool cover -func=cover.out

testacc:
	TF_ACC=1 go test -v -cover -coverprofile=cover.out -timeout 120m ./...
	go tool cover -func=cover.out

.PHONY: build install lint generate fmt test testacc