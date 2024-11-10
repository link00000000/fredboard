.PHONY: build
build : *.go
	go build -o result/fredboard

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
