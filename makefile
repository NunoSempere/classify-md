build:
	go build -o topicator

install:
	sudo cp topicator /usr/bin/topicator

run:
	go run main.go example/topics.txt example/markdown.md


