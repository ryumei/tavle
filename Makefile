#
#   Makefile for go project
#

NAME     := tavle
VERSION  := $(shell git describe --tag)
REVISION := $(shell git rev-parse --short HEAD)

SRCS    := $(shell find . -type f -name '*.go')
LDFLAGS := -ldflags="-s -w -X \"main.Version=$(VERSION)\" -X \"main.Revision=$(REVISION)\" -extldflags \"-static\""

SUBPKGS := $(shell glide novendor)

.PHONY: all 
all: test bin/$(NAME)

bin/$(NAME): $(SRCS)
	go build -a -tags netgo -installsuffix netgo $(LDFLAGS) -o bin/$(NAME)

.PHONY: clean
clean:
	rm -fr bin/*
	rm -fr vendor/*

.PHONY: test testv bench
test: deps golint
	go test $(SUBPKGS)

testv: test
	go test -v $(SUBPKGS)

bench: test
	go test -bench . ./... -benchmem

.PHONY: golint
golint:
	@for d in $(SUBPKGS); do \
	  golint $$d;\
	done

coverage.out: test
	echo 'mode: atomic' > $@ && \
	glide novendor |\
	xargs -n1 -I{} sh -c 'go test -covermode=atomic -coverprofile=coverage.tmp {} && tail -n +2 coverage.tmp >> coverage.out' && \
	rm coverage.tmp
coverage: coverage.out
	go tool cover -html=$<

.PHONY: cross-build
cross-build: test
	for os in darwin linux windows; do \
	  for arch in amd64 386; do \
	    GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo $(LDFLAGS) -o dist/$$os-$$arch/$(NAME); \
	  done; \
	done; \
	for arch in amd64 386; do \
	  GOOS=windows GOARCH=$$arch CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo $(LDFLAGS) -o dist/windows-$$arch/$(NAME).exe; \
	done


DIST_DIRS := find * -maxdepth 0 -type d -exec

.PHONY: dist
dist: cross-build
	cd dist && \
	$(DIST_DIRS) cp ../LICENSE {} \; && \
	$(DIST_DIRS) cp ../README.md  {} \; && \
	$(DIST_DIRS) cp ../tavle.tml.sample  {} \;  && \
	$(DIST_DIRS) cp -r ../public {} \;  && \
	$(DIST_DIRS) tar -zcf $(NAME)-$(VERSION)-{}.tar.gz {} \;  && \
	$(DIST_DIRS) zip -r $(NAME)-$(VERSION)-{}.zip {} \;  && \
	cd ..

.PHONY: glide
glide:
ifeq ($(shell command -v glide 2>/dev/null),)
	curl https://glide.sh/get | sh
endif

.PHONY: deps
deps: glide
	glide install update

.PHONY: dist-src
dist-src:
	git archive --format=zip -o $(NAME)-src.$(VERSION).zip HEAD

