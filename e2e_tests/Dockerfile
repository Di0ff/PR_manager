FROM golang:1.24-alpine

WORKDIR /tests

RUN apk add --no-cache ca-certificates curl

COPY go.mod go.sum* ./
RUN go mod download || true

COPY . .

CMD ["go", "test", "-v", "./..."]
