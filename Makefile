all: binaries

SHELL := /bin/bash

ALL_GO_FILES := $(shell find ./ -name '*.go') $(shell find ./ -type f -wholename '*/embedded/*')

.PHONY: binaries
binaries: binaries_ui binaries_ui-demo

.PHONY: binaries_ui
binaries_ui: bin/ui/app bin/ui/web/app.wasm bin/ui/web/main.css

.PHONY: binaries_ui-demo
binaries_ui-demo: bin/ui-demo/app bin/ui-demo/web/app.wasm bin/ui-demo/web/main.css

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

bin/ui-demo: bin
	mkdir -p $@

bin/ui-demo/app: bin/ui-demo $(ALL_GO_FILES)
	go build -o $@ ./cmd/ui-demo/...

bin/ui-demo/web: bin/ui-demo
	mkdir -p $@

bin/ui-demo/web/app.wasm: bin/ui-demo/web $(ALL_GO_FILES)
	GOARCH=wasm GOOS=js go build -o $@ ./cmd/ui-demo/...

bin/ui-demo/web/main.css: bin/ui-demo/web static/main.css
	cp static/main.css $@

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

.PHONY: run-demo
run: binaries_ui-demo
	cd bin && ./ui-demo/app

.PHONY: watch-run-demo
watch-run-demo:
	@PID=; \
	trap 'kill $$PID' TERM INT; \
	while true; do \
		$(MAKE) binaries_ui-demo; \
		ok=$$?; \
		command=$$(if [ $$ok -eq 0 ]; then echo "./app"; else echo "sleep infinity"; fi); \
		pushd bin/ui-demo; \
		$$command & PID=$$!; \
		popd; \
		inotifywait -qre close_write .; \
		kill $$PID; \
	done

.PHONY: clean
clean:
	rm -f bin/ui/app
	rm -f bin/ui/web/app.wasm
	rm -rf bin/ui/web
	rm -f bin/ui-demo/app
	rm -f bin/ui-demo/web/app.wasm
	rm -rf bin/ui-demo/web

