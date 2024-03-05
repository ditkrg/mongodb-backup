FROM golang:1.22.0-alpine as builder

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./

RUN go build -tags=nomsgpack -v -o server cmd/main.go

FROM reg.dev.krd/hub.docker/library/debian:stable-slim

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/server /app/server

RUN addgroup --group "app" --gid 1001 && adduser --uid 1001 --gid 1001 "app"

RUN chown app:app /app

USER app

ENTRYPOINT ["/app/server"]
