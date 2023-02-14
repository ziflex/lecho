default: install build

add:
	go get -u -v ${PKG}

install-tools:
	go install honnef.co/go/tools/cmd/staticcheck@latest && \
	go install golang.org/x/tools/cmd/goimports@latest


install-deps:
	go mod tidy

install: install-tools install-deps

build: lint test

test:
	go test -race ./...

lint:
	go vet ./... && \
	staticcheck -tests=false ./...

fmt:
	go fmt ./... && \
	goimports -w .