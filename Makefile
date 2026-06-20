.PHONY: build serve
build:
	go build -o spd

serve: build
	./spd serve

${V}.SILENT:
