FREDBOARD_SERVER_VERSION := $(FREDBOARD_SERVER_VERSION)
FREDBOARD_SERVER_COMMIT := $(FREDBOARD_SERVER_COMMIT)

.PHONY: build
build : *.go
	go build -ldflags "-X main.version=$(FREDBOARD_SERVER_VERSION) -X main.commit=$(FREDBOARD_SERVER_COMMIT)" -o result/fredboard

.PHONY: run
run :
	dotenv -- go run .

.PHONY: debug
debug :
	dotenv -- dlv debug .

.PHONY: test
test :
	dotenv -- go test .

.PHONY: debug-test
debug-test :
	dotenv -- dlv test .

.PHONY: dev
dev :
	dotenv -- go run ./cmd/dev
