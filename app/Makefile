default: build

$(TOOLS):
	go install $(TOOLS)

run:
	go run .

test:
	go test ./...

build:
	go build .

fmt:
	gofmt -l -s -w .

gen:
	go generate ./...