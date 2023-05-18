FROM golang:1.19 AS builder
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64
WORKDIR /build
COPY . .
RUN go build -o iptv-proxy ./cmd/iptvproxy/main.go
WORKDIR /dist
RUN cp /build/iptv-proxy .

FROM alpine
RUN apk add --no-cache curl
COPY --from=builder /dist/ /
ENTRYPOINT ["/iptv-proxy"]