FROM golang:1.21-alpine AS builder

WORKDIR /build

# Copy SDK dependency first for layer caching
COPY sdk/go/ ./sdk/go/

# Copy CLI source
COPY cli/ ./cli/

WORKDIR /build/cli

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /anyvali .

FROM scratch

COPY --from=builder /anyvali /anyvali

EXPOSE 8080

ENTRYPOINT ["/anyvali"]
CMD ["serve"]
