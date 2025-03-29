BUILD_VERSION := $(FREDBOARD_BUILD_VERSION)
BUILD_COMMIT := $(FREDBOARD_BUILD_COMMIT)

ifeq ($(OS),Windows_NT)
    SHARED_LIB_EXTENSION := .dll
else
    UNAME := $(shell uname 2>/dev/null || echo Unknown)
    
    ifeq ($(UNAME),Linux)
        SHARED_LIB_EXTENSION := .so
    else ifeq ($(UNAME),Darwin)
        SHARED_LIB_EXTENSION := .dylib
    else
        SHARED_LIB_EXTENSION :=
    endif
endif

export FREDBOARD_CONFIG := ./.env/config.json

.PHONY: default
default : fredboard

.PHONY: run
run : run-fredboard

.PHONY: debug
debug : debug-fredboard

.PHONY: all
all : fredboard audiograph

.PHONY: clean
clean :
	@rm -rf ./bin

#----------------------
# Fredboard
#----------------------

CMD_FREDBOARD = ./cmd/fredboard/

.PHONY: fredboard
fredboard : $(wildcard **/*.go) libgui
	go build -v -ldflags "-X main.buildVersion=$(BUILD_VERSION) -X main.buildCommit=$(BUILD_COMMIT)" -tags=gui,debug -o bin/fredboard $(CMD_FREDBOARD)

.PHONY: run-fredboard
run-fredboard : libgui
	go run -tags=gui,debug ./cmd/fredboard

.PHONY: debug-fredboard
debug-fredboard :
	dlv debug $(CMD_FREDBOARD)

.PHONY: libgui
libgui : $(wildcard **/*.go)
	go build -v -ldflags "-X main.buildVersion=$(BUILD_VERSION) -X main.buildCommit=$(BUILD_COMMIT)" $(if $(OS),-tags=linux$(endif)) -tags=gui,debug -buildmode=c-shared -o bin/libgui$(SHARED_LIB_EXTENSION) ./lib/gui
