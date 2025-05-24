FROM golang:1.24.3-alpine

ENV GO111MODULE=on

WORKDIR /app

# Install required packages
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main ./cmd/server

EXPOSE 8080

# Run the server
CMD ["./main", "docker"]
