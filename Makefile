build:	export GOPATH = $(realpath ./vendor/)
build:	main.go password.go;
	go build  -o build/heracles main.go password.go

clean:
	rm -rf ./build
