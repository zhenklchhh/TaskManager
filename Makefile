run:
	go get cmd/api/main.go
docker-up:
	docker-compose up -d
lint:
	golangcli-lint run