LDFLAGS += -X 'github.com/mr-stringer/punkbot/global.ReleaseVersion=$(shell git describe --tag --abbrev=0 || echo "Development")'
LDFLAGS += -X 'github.com/mr-stringer/punkbot/global.BuildTime=$(shell date +'%d-%m-%Y_%H:%M:%S')'

.PHONY: all
all: test build

test:
	go test -v

build:
	go build -ldflags="$(LDFLAGS)" -o punkbot
