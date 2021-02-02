
.PHONY: build
build:
	docker-compose up --build -d

.PHONY: run
run:
	go run main.go -session-key=sessionsssssssssssssssssssss