FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /node ./server

FROM alpine:latest
COPY --from=build /node /usr/local/bin/node
ENTRYPOINT ["node"]
