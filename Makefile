.PHONY: clean build all

clean:
		find bin -type f  -delete

build:
		GOOS=darwin go build -o bin/osx/go-cimd main.go
		GOOS=linux go build -o bin/linux/go-cimd main.go
		#GOOS=windows go build -o bin/windows/go-cimd.exe main.go

all: clean build
