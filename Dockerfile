FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/converter_server .

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /bin/converter_server /app/converter_server

ENV PORT=8080

EXPOSE 8080

CMD ["/app/converter_server"]
