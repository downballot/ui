all: binaries

SHELL := /bin/bash

ALL_GO_FILES := $(shell find ./ -name '*.go') $(shell find ./ -type f -wholename '*/embedded/*')

.PHONY: binaries
binaries: binaries_ui

.PHONY: binaries_ui
binaries_ui: bin/ui/app bin/ui/web/app.wasm bin/ui/web/main.css

bin:
	mkdir -p $@

bin/ui: bin
	mkdir -p $@

bin/ui/app: bin/ui $(ALL_GO_FILES)
	go build -o $@ ./cmd/ui/...

bin/ui/web: bin/ui
	mkdir -p $@

bin/ui/web/app.wasm: bin/ui/web $(ALL_GO_FILES)
	GOARCH=wasm GOOS=js go build -o $@ ./cmd/ui/...

bin/ui/web/main.css: bin/ui/web static/main.css
	cp static/main.css $@

bin/ui/web/robots.txt: bin/ui/web static/robots.txt
	cp static/robots.txt $@

.PHONY: run
run: binaries_ui
	cd bin && ./ui/app

.PHONY: watch-run
watch-run:
	@PID=; \
	trap 'kill $$PID' TERM INT; \
	while true; do \
		$(MAKE) binaries_ui; \
		ok=$$?; \
		command=$$(if [ $$ok -eq 0 ]; then echo "./app"; else echo "sleep infinity"; fi); \
		pushd bin/ui; \
		$$command & PID=$$!; \
		popd; \
		inotifywait -qre close_write .; \
		kill $$PID; \
	done

.PHONY: clean
clean:
	rm -rf bin

.PHONY: test
test:
	go vet ./...
	go test ./...

