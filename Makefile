.PHONY: build
build : *.go
	go build -o result/fredboard.go

.PHONY: run
run :
	go run .

