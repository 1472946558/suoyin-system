FROM golang:1.25.4-alpine AS builder

WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o /out/server ./cmd/server

FROM alpine:3.20

RUN adduser -D -H appuser
WORKDIR /app
COPY --from=builder /out/server /app/server
USER appuser
EXPOSE 8080
CMD ["/app/server"]
