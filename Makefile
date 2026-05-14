.PHONY: build run clean

build:
	go build -o converter_server .

run: build
	./converter_server

clean:
	rm -f converter_server
