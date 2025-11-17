# Multi-stage build: Alpine for compilation, distroless for runtime
# Final image: ~36MB (binary + 14MB base + certs/tzdata)

FROM golang:1.25-alpine AS builder

RUN apk add --no-cache ca-certificates tzdata git

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Static binary: no CGO, stripped symbols, reproducible builds
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-w -s" \
    -o /build/cruder \
    ./cmd/main.go

# Runtime: distroless provides CA certs + tzdata with minimal attack surface
FROM gcr.io/distroless/static-debian12

# Copy timezone data from builder (for accurate timestamps in logs)
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo


COPY --from=builder /build/cruder /cruder

# For CI/CD: Container can run migrations before starting the app
COPY migrations /migrations

EXPOSE 8080

# Non-root user for security
USER nonroot:nonroot

ENV GIN_MODE=release

ENTRYPOINT ["/cruder"]


# while developing locally, the image size is 36.4MB.