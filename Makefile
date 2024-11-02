.PHONY: build
build : *.go
	go build -o result/fredboard.go

.PHONY: run
run :
	dotenv -- go run .

.PHONY: debug
debug :
	dotenv -- dlv debug .

