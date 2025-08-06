SWAG=$(shell go env GOPATH)/bin/swag

.PHONY: swagger build run clean

swagger:
	$(SWAG) init -g cmd/main.go

build:
	go build -o bin/app cmd/main.go

run:
	go run cmd/main.go

clean:
	rm -rf bin docs/swagger