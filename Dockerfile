FROM golang:1.21.3-alpine3.17 as builder

# Ca-certificates are required to call HTTPS endpoints.
RUN apk update && apk add --no-cache ca-certificates tzdata alpine-sdk bash && update-ca-certificates

# Create appuser
RUN adduser -D -g '' appuser
WORKDIR /app
# Copy go mod and sum files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# compile the app
COPY cmd cmd
COPY internal internal

# bundle version into binary if specified in build-args, dev otherwise.
ARG version=dev
# Compile statically
RUN CGO_ENABLED=0 go build -ldflags "-w -extldflags '-static' -X internal/app.Version=${version}" -o /proxy cmd/proxy/proxy.go

FROM scratch

COPY --from=builder /proxy /proxy
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 6001

HEALTHCHECK  --interval=5s --start-period=5s --timeout=3s CMD ["/proxy", "healthcheck"]
CMD ["/proxy"]
