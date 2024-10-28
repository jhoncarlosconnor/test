FROM golang:1.22.5 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 80
CMD ["sh", "-c", "URL=redis-19606.c1.us-central1-2.gce.redns.redis-cloud.com:19606 PASSWORD=duVLW1QEN3K6tHWjOUOGSnsTRuB5TdB7 ./main"]
