all: binaries

SHELL := /bin/bash

ALL_GO_FILES := $(shell find ./ -name '*.go')

.PHONY: binaries
binaries: bin/ui bin/web/app.wasm bin/web/main.css

bin:
	mkdir -p $@

bin/ui: bin $(ALL_GO_FILES)
	go build -o $@ ./cmd/ui/...

bin/web: bin
	mkdir -p $@

bin/web/app.wasm: bin/web $(ALL_GO_FILES)
	GOARCH=wasm GOOS=js go build -o $@ ./cmd/ui/...

bin/web/main.css: bin/web static/main.css
	cp static/main.css $@

bin/web/robots.txt: bin/web static/robots.txt
	cp static/robots.txt $@

.PHONY: run
run: binaries
	cd bin && ./ui

.PHONY: watch-run
watch-run:
	@PID=; \
	trap 'kill $$PID' TERM INT; \
	while true; do \
		$(MAKE) binaries; \
		ok=$$?; \
		command=$$(if [ $$ok -eq 0 ]; then echo "./ui"; else echo "sleep infinity"; fi); \
		pushd bin; \
		$$command & PID=$$!; \
		popd; \
		inotifywait -qre close_write .; \
		kill $$PID; \
	done


.PHONY: clean
clean:
	rm -f bin/ui
	rm -f bin/web/app.wasm
	rm -rf bin/web
	rm -rf 
