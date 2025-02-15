FROM golang:1.24-alpine AS builder

WORKDIR /usr/local/src/chatmailer

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o ./chatmailer -ldflags="-s -w" ./cmd/chatmailer/


FROM golang:1.24-alpine AS runner

WORKDIR /usr/local/src/chatmailer

COPY --from=builder /usr/local/src/chatmailer/chatmailer /usr/local/bin/chatmailer

EXPOSE 8081

ENTRYPOINT ["/usr/local/bin/chatmailer"]
