FROM golang:1.24.6-alpine

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o server ./cmd/main.go

COPY environment/.env .env

EXPOSE 8080

CMD ["./server"]
