FROM golang:1.27.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o ruangku .

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates && touch .env

COPY --from=builder /app/ruangku .

EXPOSE 3000

CMD ["./ruangku"]


