FROM golang:1.21-alpine3.17 as builder

# Ca-certificates are required to call HTTPS endpoints.
RUN apk update && apk add --no-cache ca-certificates tzdata alpine-sdk bash && update-ca-certificates

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-extldflags=-static" -o /proxy cmd/proxy/proxy.go

FROM scratch

COPY --from=builder /proxy /proxy
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 80

CMD ["/proxy"]
