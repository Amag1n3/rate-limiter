FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/rate-limiter .

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /out/rate-limiter /app/rate-limiter

EXPOSE 8080

ENV PORT=8080

ENTRYPOINT ["/app/rate-limiter"]
