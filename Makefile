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
	@FREDBOARD_CONFIG=./.env/config.json go run $(CMD_FREDBOARD)

.PHONY: debug-fredboard
debug-fredboard :
	FREDBOARD_CONFIG=./.env/config.json dlv debug $(CMD_FREDBOARD)

#----------------------
# Audio Graph
#----------------------

CMD_AUDIOGRAPH = ./cmd/audiograph/

.PHONY: audiograph
audiograph : $(wildcard **/*.go)
	go build -v -ldflags "-X main.buildVersion=$(BUILD_VERSION) -X main.buildCommit=$(BUILD_COMMIT)" -o bin/audiograph $(CMD_AUDIOGRAPH)

.PHONY: run-audiograph
run-audiograph :
	@FREDBOARD_CONFIG=./.env/config.json go run $(CMD_AUDIOGRAPH)

.PHONY: debug-audiograph
debug-audiograph :
	FREDBOARD_CONFIG=./.env/config.json dlv debug $(CMD_AUDIOGRAPH)

#----------------------
# Youtube Downloader
#----------------------

CMD_YOUTUBE_DOWNLOADER = ./cmd/youtube_downloader/

.PHONY: youtube-downloader
youtube-downloader : $(wildcard **/*.go)
	go build -v -ldflags "-X main.buildVersion=$(BUILD_VERSION) -X main.buildCommit=$(BUILD_COMMIT)" -o bin/youtube-downloader $(CMD_YOUTUBE_DOWNLOADER)

.PHONY: run-youtube-downloader
run-youtube-downloader :
	@FREDBOARD_CONFIG=./.env/config.json go run $(CMD_YOUTUBE_DOWNLOADER)

.PHONY: debug-youtube-downloader
debug-youtube-downloader :
	FREDBOARD_CONFIG=./.env/config.json dlv debug $(CMD_YOUTUBE_DOWNLOADER)

#----------------------
# Parallel Audio Graph
#----------------------

CMD_AUDIOGRAPH_PARALLEL = ./cmd/audiograph_parallel/

.PHONY: audiograph-parallel
audiograph-parallel : $(wildcard **/*.go)
	go build -v -ldflags "-X main.buildVersion=$(BUILD_VERSION) -X main.buildCommit=$(BUILD_COMMIT)" -o bin/audiograph-parallel $(CMD_AUDIOGRAPH_PARALLEL)

.PHONY: run-audiograph-parallel
run-audiograph-parallel :
	@FREDBOARD_CONFIG=./.env/config.json go run $(CMD_AUDIOGRAPH_PARALLEL)

.PHONY: debug-audiograph-parallel
debug-audiograph-parallel :
	FREDBOARD_CONFIG=./.env/config.json dlv debug $(CMD_AUDIOGRAPH_PARALLEL)
