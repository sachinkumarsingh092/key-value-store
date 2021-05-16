.PHONY: runserver buildclient

runserver:
	go run ./cmd/server/main.go

buildclient:
	go build ./cmd/client
