SHELL := /bin/bash

# The name of the executable (default is current directory name)
TARGET := $(shell echo $${PWD\#\#*/})
BIN_FOLDER=bin
.DEFAULT_GOAL: $(TARGET)

# These will be provided to the target
VERSION := 1.1.0 
BUILD := `git rev-parse HEAD`
DATE := `date +"%d.%B.%Y-%T"`

# Use linker flags to provide version/build settings to the target
LDFLAGS=-ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD) -X=main.BuildDate="$(DATE)""

# go source files, ignore vendor directory
SRC := main.go 

.PHONY: all build clean install uninstall fmt simplify check run

all: check install

build:  
	@go build -o $(BIN_FOLDER)/$(TARGET) ${LDFLAGS} $(SRC) 

clean:
	@rm -f $(TARGET)

install:
	@go install $(LDFLAGS)

test: 
	@go test -run ''

uninstall: clean
	@rm -f $$(which ${TARGET})

fmt:
	@gofmt -l -w $(SRC)

simplify:
	@gofmt -s -l -w $(SRC)

check:
	@test -z $(shell gofmt -l main.go | tee /dev/stderr) || echo "[WARN] Fix formatting issues with 'make fmt'"
	@for d in $$(go list ./... | grep -v /vendor/); do golint $${d}; done
	@go tool vet ${SRC}

run: build 
	./$(BIN_FOLDER)/$(TARGET) --config ./conf.yaml 
run2: build 
	./$(BIN_FOLDER)/$(TARGET) --config ./conf.yaml --logtostderr -v 3 
run3: build 
	./$(BIN_FOLDER)/$(TARGET) --config ./conf2.yaml 

