# Stage 1 - Builder stage
FROM golang:1.21 AS builder
WORKDIR /go/src
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o app .

# Stage 2 - Deployer stage
FROM alpine:latest
ENV TZ=Asia/Bangkok
RUN apk add --no-cache tzdata
WORKDIR /app
COPY --from=builder /go/src/app .
EXPOSE 3000
CMD ["./app"]
