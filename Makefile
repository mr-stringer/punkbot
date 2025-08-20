LDFLAGS += -X 'main.ReleaseVersion=$(shell git describe --tag --exact-match --abbrev=0 || echo "Development")'
LDFLAGS += -X 'main.BuildTime=$(shell date +'%d-%m-%Y_%H:%M:%S')'

.PHONY: all
all: test build

test:
	go test -v

build:
	go build -ldflags="$(LDFLAGS)" -o punkbot
