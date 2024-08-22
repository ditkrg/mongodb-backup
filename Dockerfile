FROM golang:1.23.0-alpine as builder

WORKDIR /app

COPY go.* ./

RUN go mod download

COPY . ./

RUN go build -tags=nomsgpack -v -o /app/cli main.go

FROM scratch

COPY --from=builder --chown=1001:1001 /app /app

USER 1001

ENTRYPOINT ["/app/cli"]
