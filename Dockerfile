FROM golang:1.23-alpine3.20 as builder

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
RUN CGO_ENABLED=0 go build -ldflags "-w -extldflags '-static'" -o /healthcheck cmd/healthcheck/healthcheck.go

FROM scratch

COPY --from=builder /proxy /proxy
COPY --from=builder /healthcheck /healthcheck
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 4321
HEALTHCHECK --interval=1s --timeout=1s --start-period=2s --retries=3 CMD [ "/healthcheck" ]

CMD ["/proxy"]
