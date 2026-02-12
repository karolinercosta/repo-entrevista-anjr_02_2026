FROM golang:1.20-alpine AS builder

WORKDIR /src

# Cache modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags='-s -w' -o /app/taskapi .

# Final image
FROM alpine:3.18
RUN addgroup -S app && adduser -S app -G app
COPY --from=builder /app/taskapi /usr/local/bin/taskapi
USER app
EXPOSE 8080
ENV MONGO_URI="mongodb://mongo:27017"
CMD ["/usr/local/bin/taskapi"]
