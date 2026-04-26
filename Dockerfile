FROM golang:1.26-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o /leaderboard ./cmd/leaderboard/

# ---
FROM alpine:3.21

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /leaderboard .

EXPOSE 8080

VOLUME ["/app/data"]

ENV LEADERBOARD_DB_PATH=/app/data/leaderboard.db

ENTRYPOINT ["./leaderboard"]
CMD ["serve"]
