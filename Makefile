build:	export GOPATH = $(realpath ./vendor/)
build:	main.go;
	go build  -o build/heracles main.go

clean:
	rm -rf ./build
