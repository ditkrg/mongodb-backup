FROM golang:1.22.0-alpine as builder

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./

RUN go build -tags=nomsgpack -v -o server cmd/main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder app/server .
USER 65532:65532

ENTRYPOINT ["/server"]
