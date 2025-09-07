export OTEL_EXPORTER_OTLP_ENDPOINT := http://localhost:4318
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
	@rm -rf bin/

#----------------------
# Fredboard Server
#----------------------

CMD_FREDBOARD = ./cmd/fredboard_server/

.PHONY: run-fredboard
run-fredboard :
	@go run $(CMD_FREDBOARD)

