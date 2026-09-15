FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
# with_utls — reality; with_quic — hysteria2/tuic; with_grpc — gRPC-transport.
RUN CGO_ENABLED=0 go build -trimpath -tags "with_utls,with_quic,with_grpc" -ldflags "-s -w -X github.com/sagernet/sing-box/constant.Version=1.14.0 -X main.version=$VERSION" -o /node ./server

FROM alpine:latest
COPY --from=build /node /usr/local/bin/node
ENTRYPOINT ["node"]
