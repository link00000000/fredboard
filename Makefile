BUILD_VERSION := $(FREDBOARD_BUILD_VERSION)
BUILD_COMMIT := $(FREDBOARD_BUILD_COMMIT)

export FREDBOARD_CONFIG := "./.env/config.json"

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
	@rm -rf bin/

#----------------------
# Fredboard Server
#----------------------

CMD_FREDBOARD = ./cmd/fredboard_server/

.PHONY: fredboard
fredboard : $(wildcard **/*.go)
	go build -v -ldflags "-X main.buildVersion=$(BUILD_VERSION) -X main.buildCommit=$(BUILD_COMMIT)" -o bin/fredboard-server $(CMD_FREDBOARD)

.PHONY: run-fredboard
run-fredboard :
	@go run $(CMD_FREDBOARD)

.PHONY: debug-fredboard
debug-fredboard :
	dlv debug $(CMD_FREDBOARD)

#----------------------
# Audio Graph GUI
#----------------------

CMD_AUDIOGRAPH_GUI = ./cmd/audiograph-gui/

.PHONY: audiograph-gui
audiograph-gui : $(wildcard **/*.go)
	go build -v -ldflags "-X main.buildVersion=$(BUILD_VERSION) -X main.buildCommit=$(BUILD_COMMIT)" -o bin/audiograph $(CMD_AUDIOGRAPH_GUI)

.PHONY: run-audiograp-gui
run-audiograph-gui :
	@go run $(CMD_AUDIOGRAPH_GUI)

.PHONY: debug-audiograph-gui
debug-audiograph-gui :
	dlv debug $(CMD_AUDIOGRAPH_GUI)

#----------------------
# Youtube Downloader
#----------------------

CMD_YOUTUBE_DOWNLOADER = ./cmd/youtube_downloader/

.PHONY: youtube-downloader
youtube-downloader : $(wildcard **/*.go)
	go build -v -ldflags "-X main.buildVersion=$(BUILD_VERSION) -X main.buildCommit=$(BUILD_COMMIT)" -o bin/youtube-downloader $(CMD_YOUTUBE_DOWNLOADER)

.PHONY: run-youtube-downloader
run-youtube-downloader :
	@go run $(CMD_YOUTUBE_DOWNLOADER)

.PHONY: debug-youtube-downloader
debug-youtube-downloader :
	dlv debug $(CMD_YOUTUBE_DOWNLOADER)

#----------------------
# Parallel Audio Graph
#----------------------

CMD_AUDIOGRAPH_PARALLEL = ./cmd/audiograph_parallel/

.PHONY: audiograph-parallel
audiograph-parallel : $(wildcard **/*.go)
	go build -v -ldflags "-X main.buildVersion=$(BUILD_VERSION) -X main.buildCommit=$(BUILD_COMMIT)" -o bin/audiograph-parallel $(CMD_AUDIOGRAPH_PARALLEL)

.PHONY: run-audiograph-parallel
run-audiograph-parallel :
	@go run $(CMD_AUDIOGRAPH_PARALLEL)

.PHONY: debug-audiograph-parallel
debug-audiograph-parallel :
	dlv debug $(CMD_AUDIOGRAPH_PARALLEL)
