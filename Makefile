BUILD_VERSION := $(FREDBOARD_BUILD_VERSION)
BUILD_COMMIT := $(FREDBOARD_BUILD_COMMIT)

# tags:
# 	- gui_glfw	: enable gui with GLFW backend
# 	- gui_sdl		: enable gui with SDL backend
TAGS := -tags=gui_glfw

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
	rm -rf bin/

#----------------------
# Fredboard
#----------------------

CMD_FREDBOARD = ./cmd/fredboard/

.PHONY: fredboard
fredboard : $(wildcard **/*.go)
	go build -v -ldflags "-X main.buildVersion=$(BUILD_VERSION) -X main.buildCommit=$(BUILD_COMMIT)" $(TAGS) -o bin/fredboard $(CMD_FREDBOARD)

.PHONY: run-fredboard
run-fredboard :
	go run $(TAGS) $(CMD_FREDBOARD)

.PHONY: debug-fredboard
debug-fredboard :
	dlv debug $(CMD_FREDBOARD)

