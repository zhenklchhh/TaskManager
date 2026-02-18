run:
	go get cmd/api/main.go
	go get cmd/worker/worker.go
docker-up:
	docker-compose up -d
lint:
	golangcli-lint run