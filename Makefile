VERSION=0.0.6
LDFLAGS=-ldflags "-w -s -X main.version=${VERSION}"
all: mackerel-plugin-log-counter

.PHONY: mackerel-plugin-linux-process-status

mackerel-plugin-log-counter: main.go
	go build $(LDFLAGS) -o mackerel-plugin-log-counter

linux: main.go
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o mackerel-plugin-log-counter

fmt:
	go fmt ./...

check:
	go test -v ./...

clean:
	rm -rf mackerel-plugin-log-counter

