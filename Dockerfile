FROM golang:1.26-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download || true

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /gemini-web2api .

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /gemini-web2api /gemini-web2api
EXPOSE 8081

ENTRYPOINT ["/gemini-web2api"]
