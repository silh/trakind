FROM golang:1.19-alpine as builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY ./ ./
RUN GOOS=linux GOARCH=arm64 go build ./cmd/bot

FROM alpine

ENV TELEGRAM_API_KEY=""
ENV UPDATE_INTERVAL="1m"

COPY --from=builder /app/bot /bot

ENTRYPOINT ["/bot"]
