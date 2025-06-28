BUILD_VERSION := $(or $(GAPS_BUILD_VERSION), "dev")
BUILD_COMMIT := $(or $(GAPS_BUILD_COMMIT), "dev")

.PHONY: default
default : bin/gaps

.PHONY: clean
clean :
	@rm -rf bin/

#----------------------
# GAPS
#----------------------

CMD_GAPS = .

bin/gaps : $(wildcard **/*.go)
	go build -v -ldflags "-X main.buildVersion=$(BUILD_VERSION) -X main.buildCommit=$(BUILD_COMMIT)" -o $@ $(CMD_GAPS)

