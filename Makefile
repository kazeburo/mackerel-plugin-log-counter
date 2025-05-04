VERSION=0.0.9
LDFLAGS=-ldflags "-w -s -X main.version=${VERSION}"
all: mackerel-plugin-log-counter

.PHONY: mackerel-plugin-linux-process-status

mackerel-plugin-log-counter: cmd/mackerel-plugin-log-counter/*.go
	go build $(LDFLAGS) -o mackerel-plugin-log-counter cmd/mackerel-plugin-log-counter/*.go

linux: cmd/mackerel-plugin-log-counter/*.go
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o mackerel-plugin-log-counter cmd/*.go

fmt:
	go fmt ./...

check:
	go test -v ./...

clean:
	rm -rf mackerel-plugin-log-counter

