FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/main.go

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates bash curl

COPY --from=builder /app/server /app/server

COPY db/scripts ./db/scripts

EXPOSE 8080

CMD ["/app/server"]
