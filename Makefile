.PHONY: up down logs proto

up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       proto/node/node.proto
