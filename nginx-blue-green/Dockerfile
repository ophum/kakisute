FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod /app/go.mod

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/app ./

FROM alpine

WORKDIR /app

COPY --from=builder /app/app /app/app

ENTRYPOINT ["/app/app"]