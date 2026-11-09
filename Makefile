.PHONY: build run clean

build:
	go build -o cmd/mandarin/mandarin ./cmd/mandarin

run: build
	cd /Users/cj/Workspace/mandarin-wiki && ./cmd/mandarin/mandarin

clean:
	rm -f cmd/mandarin/mandarin
