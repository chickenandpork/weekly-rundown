# syntax=docker/dockerfile:1

FROM golang:1.25-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /weekly-rundown ./cmd/server/
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /weekly-rundown-health ./cmd/healthcheck/

FROM cgr.dev/chainguard/static:latest
COPY --from=builder /weekly-rundown /weekly-rundown
COPY --from=builder /weekly-rundown-health /weekly-rundown-health
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  /weekly-rundown-health
ENTRYPOINT ["/weekly-rundown"]
