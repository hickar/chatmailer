FROM golang:1.21.3-alpine as build

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o ./chatmailer ./cmd/chatmailer/

FROM golang:1.21.3-alpine

WORKDIR /app

RUN go install github.com/cosmtrek/air@latest

COPY --from=build /app/config.yaml /app/config.yaml
COPY --from=build /app/chatmailer /app/chatmailer
COPY --from=build /app/.air.toml /app/.air.toml

EXPOSE 8081

CMD ["air", "-c", ".air.toml"]
