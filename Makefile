GO		= go
DEP		= dep
GIT     = git
GOLINT  = golint
GREP    = grep

PACKAGE	= github.com/zekroTJA/vplan2019
GOPATH	= $(CURDIR)/.gopath
WDIR	= $(GOPATH)/src/$(PACKAGE)

BINNAME	= vplan2019_server
BINLOC	= $(CURDIR)

ifeq ($(OS),Windows_NT)
	EXTENSION = .exe
endif

BIN		= $(BINLOC)/$(BINNAME)$(EXTENSION)

TAG		= $(shell $(GIT) describe --tags)
COMMIT	= $(shell $(GIT) rev-parse HEAD)
GOVERS  = $(shell $(GO) version | sed -e 's/ /_/g')

.PHONY: _make deps cleanup _finish run lint offline release

_make: $(WDIR) deps $(BIN) cleanup _finish

offline: $(WDIR) $(BIN) cleanup _finish

$(WDIR):
	@echo [ INFO ] creating working directory '$@'...
	mkdir -p $@
	cp -R $(CURDIR)/* $@/

$(BIN):
	@echo [ INFO ] building binary '$(BIN)'...
	(env GOPATH=$(GOPATH) ${ARGS} $(GO) build -v -o $@ -ldflags "\
		-X $(PACKAGE)/internal/ldflags.AppVersion=$(TAG) \
		-X $(PACKAGE)/internal/ldflags.AppCommit=$(COMMIT) \
		-X $(PACKAGE)/internal/ldflags.GoVersion=$(GOVERS) \
		-X $(PACKAGE)/internal/ldflags.Release=TRUE" \
		$(WDIR)/cmd/server)

release: $(WDIR) $(BIN) deps cleanup
	@echo [ INFO ] Creating release...
	mkdir $(CURDIR)/release
	mv -f $(BIN) $(CURDIR)/release
	cp -f -R $(CURDIR)/web $(CURDIR)/release/web

deps:
	@echo [ INFO ] getting dependencies...	
	cd $(WDIR) && \
		$(DEP) ensure -v

cleanup:
	@echo [ INFO ] cleaning up...
	rm -r -f $(GOPATH)

_finish:
	@echo ------------------------------------------------------------------------------
	@echo [ INFO ] Build successful.
	@echo [ INFO ] Your build is located at '$(BIN)'

run:
	@echo [ INFO ] Debug running...
	(env GOPATH=$(CURDIR)/../../../.. $(GO) run -v ./cmd/server -c $(CURDIR)/config/config.yml ${ARGS})

lint:
	$(GOLINT) ./... | $(GREP) -v vendor