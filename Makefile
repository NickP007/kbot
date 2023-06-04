ifeq ($(strip $(APP)),)
APP := $(shell basename $(shell git remote get-url origin) | sed 's/\.git$$//')
endif
ifeq ($(strip $(REGISTRY)),)
REGISTRY := docker.io
endif
ifeq ($(strip $(REPOSITORY)),)
REPOSITORY := nickp6
endif
VERSION := $(shell git describe --tags --abbrev=0)-$(shell git rev-parse --short HEAD)
ifeq ($(strip $(TARGETOS)),)
TARGETOS := linux
endif
ifeq ($(strip $(TARGETARCH)),)
TARGETARCH := $(shell dpkg --print-architecture)
endif

format:
	gofmt -s -w ./

lint:
	golint

test:
	go test -v

get:
	go get

build: format get
	CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -v -o kbot -ldflags "-X="github.com/NickP007/kbot/cmd.AppVersionNum=${VERSION}

image:
	docker build . -t ${REGISTRY}/${REPOSITORY}/${APP}:${VERSION}-${TARGETOS}-${TARGETARCH} --build-arg TARGETOS=${TARGETOS} --build-arg TARGETARCH=${TARGETARCH} --build-arg VERSION=${VERSION}

push:
	docker push ${REGISTRY}/${REPOSITORY}/${APP}:${VERSION}-${TARGETOS}-${TARGETARCH}

clean:
	rm -rf kbot
