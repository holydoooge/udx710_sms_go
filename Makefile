BINARY=build/main
SRC=main.go db.go ofono.go sysinfo.go network.go
GOOS=linux
GOARCH=arm64

all: build

build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BINARY) $(SRC)
	aarch64-linux-gnu-strip $(BINARY)
	upx $(BINARY)

run: build
	./$(BINARY)

clean:
	rm -rf $(BINARY)

.PHONY: all build run clean
