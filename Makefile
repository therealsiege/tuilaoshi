.PHONY: build run test fmt check lint golden recordings launcher demonart clean

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

# Rewrite screen snapshots (testdata/*.golden) after an intended UI change; review the diff.
golden:
	UPDATE_GOLDEN=1 go test ./internal/ui/...

# Record docs/recordings/*.gif from docs/tapes (needs vhs: brew install vhs).
recordings:
	@for t in docs/tapes/*.tape; do vhs "$$t" || exit 1; done

# Put TuiLaoShi.app on the Desktop (double-click to start the tutor).
launcher:
	./scripts/install-launcher.sh

# Paint demon portraits with Gemini Flash Image (GEMINI_API_KEY in ~/.config/tuilaoshi/env).
# Pass ARGS, e.g. make demonart ARGS="-only 白骨精 -force".
demonart:
	zsh -c 'source ~/.config/tuilaoshi/env 2>/dev/null; go run ./cmd/demonart $(ARGS)'

clean:
	rm -f cmd/mandarin/mandarin
