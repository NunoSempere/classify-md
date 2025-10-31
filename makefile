build:
	go build -o classify

install:
	sudo cp classify /usr/bin/classify

run:
	go run main.go example/topics.txt example/markdown.md


