.PHONY: runserver buildclient clean

NAME=client

runserver:
	go run ./cmd/server/main.go

buildclient:
	go build ./cmd/client

clean:
	rm ${NAME}