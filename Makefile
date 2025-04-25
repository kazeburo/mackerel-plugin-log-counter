VERSION=0.0.6
LDFLAGS=-ldflags "-w -s -X main.version=${VERSION}"
all: mackerel-plugin-log-counter

.PHONY: mackerel-plugin-linux-process-status

mackerel-plugin-log-counter: cmd/mackerel-plugin-log-counter/main.go cmd/mackerel-plugin-log-counter/parser.go
	go build $(LDFLAGS) -o mackerel-plugin-log-counter cmd/mackerel-plugin-log-counter/main.go cmd/mackerel-plugin-log-counter/parser.go

linux: cmd/mackerel-plugin-log-counter/main.go cmd/mackerel-plugin-log-counter/parser.go
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o mackerel-plugin-log-counter cmd/mackerel-plugin-log-counter/main.go cmd/mackerel-plugin-log-counter/parser.go

fmt:
	go fmt ./...

check:
	go test -v ./...

clean:
	rm -rf mackerel-plugin-log-counter

