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

clean:
	{ \
		read -p "ok? (y/N): " yn; \
		case "$$yn" in \
		[yY]*) rm -rf app out/* ../db/*;; \
		*) echo "abort";; \
		esac; \
	}