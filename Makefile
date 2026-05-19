BINARY  := pydict2json
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.version=$(VERSION)

# Cross-compilation targets: OS/ARCH
TARGETS := \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64 \
	windows/arm64

# Derive output binary name (append .exe for windows)
out_name = $(if $(filter windows/%,$(1)),$(BINARY).exe,$(BINARY))

# Derive output directory
out_dir = dist/$(subst /,-,$(1))

.PHONY: all build clean install $(TARGETS) dist

all: build

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

install: build
	cp $(BINARY) $(GOPATH)/bin/$(BINARY)

# Single cross-compile target, e.g. make linux/amd64
$(TARGETS):
	@mkdir -p $(call out_dir,$@)
	GOOS=$(word 1,$(subst /, ,$@)) \
	GOARCH=$(word 2,$(subst /, ,$@)) \
	go build -ldflags "$(LDFLAGS)" -o $(call out_dir,$@)/$(call out_name,$@) .

# Build all targets
dist: $(TARGETS)

clean:
	rm -rf dist $(BINARY)
