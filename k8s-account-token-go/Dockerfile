FROM golang:1.26-alpine AS builder

WORKDIR /work
COPY go.mod go.sum .

RUN go mod download

COPY . .

RUN go build -o app .

FROM alpine

COPY --from=builder /work/app /app

ENTRYPOINT ["/app"]