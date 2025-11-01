# Stage 1: Build the Go application
FROM golang:latest AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Stage 2: Certificates
FROM alpine:latest AS certs
RUN apk add --no-cache ca-certificates

# Stage 3: Create the scratch image
FROM scratch
COPY --from=builder /app/main /
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
CMD [ "/main" ]
