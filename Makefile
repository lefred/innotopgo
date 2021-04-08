.DEFAULT_GOAL := build

# format code
fmt:
	go fmt ./...
.PHONY:fmt

# check code for best practices
lint: fmt
	golint ./...
.PHONY:lint

# Check code for correctness
vet: fmt
	go vet ./...
.PHONY:vet

# Build executable in current OS
build: vet
	go build .
.PHONY:build

# Build Linux executable
genlinux: vet
	CGO_ENABLED=0 go build -o build/innotopgo-linux_static .
.PHONY:genlinux

# Build Windows executable
genwin: vet
	CGO_ENABLED=0 GOOS=windows go build -o build/innotopgo-win.exe .
.PHONY:genwin

# Build MacOS executable
genmac: vet
	CGO_ENABLED=0 GOOS=darwin go build -o build/innotopgo-macos .
.PHONY:genmac

# Build all executables
genall: genmac genwin genlinux
