# Stage 1: Build the Go application
FROM golang:latest AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Stage 2: Create the scratch image
FROM scratch
COPY --from=builder /app/main /
CMD [ "/main" ]
