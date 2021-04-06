.DEFAULT_GOAL := build

fmt:
	go fmt ./...
.PHONY:fmt

lint: fmt
	golint ./...
.PHONY:lint

vet: fmt
	go vet ./...
.PHONY:vet

build: vet
	go build .
.PHONY:build

genlinux: vet
	CGO_ENABLED=0 go build -o build/innotopgo-linux_static .
.PHONY:genlinux

genwin: vet
	CGO_ENABLED=0 GOOS=windows go build -o build/innotopgo-win.exe .
.PHONY:genwin

genmac: vet
	CGO_ENABLED=0 GOOS=darwin go build -o build/innotopgo-macos .
.PHONY:genmac

genall: genmac genwin genlinux
