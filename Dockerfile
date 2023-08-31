FROM golang:1.20-alpine3.17 as builder

# Ca-certificates are required to call HTTPS endpoints.
RUN apk update && apk add --no-cache ca-certificates tzdata alpine-sdk bash && update-ca-certificates

WORKDIR /app
COPY . .

# bundle version into binary if specified in build-args, dev otherwise.
ARG version=dev
# Compile statically
RUN CGO_ENABLED=0 go build -ldflags "-w -extldflags '-static' -X internal/app.Version=${version}" -o /proxy cmd/proxy/proxy.go

FROM scratch

COPY --from=builder /proxy /proxy
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 80

CMD ["/proxy"]
