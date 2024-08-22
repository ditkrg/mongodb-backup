FROM golang:1.22.0-alpine as builder

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./

RUN go build -tags=nomsgpack -v -o cli main.go

FROM reg.dev.krd/hub.docker/library/debian:stable-slim

COPY --from=builder /app/cli /app/cli

RUN addgroup --group "app" --gid 1001 && adduser --uid 1001 --gid 1001 "app"

RUN chown app:app /app

USER app

ENTRYPOINT ["/app/cli"]
