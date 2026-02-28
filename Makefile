run-server:
	go run cmd/api/main.go

run-worker:
	go run cmd/worker/worker.go
docker-up:
	docker-compose up -d
lint:
	golangcli-lint run

migrate-up:
	migrate -database postgres://Lenovo:123456@localhost:5432/taskmanager-db?sslmode=disable -path migrations up

migrate-down:
	migrate -database postgres://Lenovo:123456@localhost:5432/taskmanager-db?sslmode=disable -path migrations down