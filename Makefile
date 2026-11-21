.PHONY: build run test fmt check lint clean

GOLANGCI_LINT = go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0

build:
	go build -o cmd/mandarin/mandarin ./cmd/mandarin

# Run from the repo root so the app finds ./wiki.
run: build
	./cmd/mandarin/mandarin

test:
	go test ./...

fmt:
	gofmt -w cmd internal

# check runs everything CI runs.
check:
	@unformatted=$$(gofmt -l cmd internal); \
	if [ -n "$$unformatted" ]; then \
		echo "These files are not gofmt-formatted. Run 'make fmt':"; echo "$$unformatted"; exit 1; \
	fi
	./scripts/check-file-length.sh
	go vet ./...
	go build ./...
	go test ./...

lint:
	$(GOLANGCI_LINT) run ./...

clean:
	rm -f cmd/mandarin/mandarin
