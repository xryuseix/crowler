default: build

$(TOOLS):
	go install $(TOOLS)

run:
	trap 'exit 0' SIGINT; go run .

test:
	go test ./...

build:
	go build .

fmt:
	gofmt -l -s -w .

gen:
	go generate ./...