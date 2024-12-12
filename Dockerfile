FROM golang:alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o server

FROM alpine

WORKDIR /app

COPY --from=builder /build/server /app/server

COPY .env /app/.env

EXPOSE 8080

CMD ["./server"]