.PHONY: up down logs proto certs test

certs:
	@mkdir -p certs
	@if [ ! -s certs/cert.pem ] || [ ! -s certs/key.pem ]; then \
		umask 077; openssl req -x509 -newkey rsa:2048 -nodes -days 365 \
		-subj '/CN=node/O=CascadiaLabs' -addext 'subjectAltName=DNS:node,DNS:localhost,IP:127.0.0.1' \
		-keyout certs/key.pem -out certs/cert.pem; \
	fi

up: certs
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

# with_utls/with_quic — те же теги, что и в Dockerfile: реальность и QUIC-протоколы.
test:
	go test -tags "with_utls,with_quic" ./...

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       proto/node/node.proto
