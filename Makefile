BUILD_VERSION := $(FREDBOARD_BUILD_VERSION)
BUILD_COMMIT := $(FREDBOARD_BUILD_COMMIT)

.PHONY: default
default : fredboard

.PHONY: run
run : run-fredboard

.PHONY: debug
debug : debug-fredboard

.PHONY: all
all : fredboard audiograph

#----------------------
# Fredboard Server
#----------------------

CMD_FREDBOARD = ./cmd/fredboard_server/

.PHONY: fredboard
fredboard : $(wildcard **/*.go)
	go build -v -ldflags "-X main.buildVersion=$(BUILD_VERSION) -X main.buildCommit=$(BUILD_COMMIT)" -o bin/fredboard-server $(CMD_FREDBOARD)

.PHONY: run-fredboard
run-fredboard :
	dotenv -- go run $(CMD_FREDBOARD)

.PHONY: debug-fredboard
debug-fredboard :
	dotenv -- dlv debug $(CMD_FREDBOARD)

#----------------------
# Audio Graph
#----------------------

CMD_AUDIOGRAPH = ./cmd/audiograph/

.PHONY: audiograph
audiograph : $(wildcard **/*.go)
	go build -v -ldflags "-X main.buildVersion=$(BUILD_VERSION) -X main.buildCommit=$(BUILD_COMMIT)" -o bin/audiograph-server $(CMD_AUDIOGRAPH)

.PHONY: run-audiograph
run-audiograph :
	dotenv -- go run $(CMD_AUDIOGRAPH)

.PHONY: debug-audiograph
debug-audiograph :
	dotenv -- dlv debug $(CMD_AUDIOGRAPH)
